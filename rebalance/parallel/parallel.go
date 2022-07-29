package parallel

import (
	"circular/graph"
	"circular/node"
	rebalance2 "circular/rebalance"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"github.com/elementsproject/glightning/jrpc2"
	"github.com/gammazero/deque"
	"sync"
	"time"
)

type RebalanceParallel struct {
	InScid              string                       `json:"inscid"`
	Amount              uint64                       `json:"amount,omitempty"`
	MaxPPM              uint64                       `json:"maxppm,omitempty"`
	Splits              int                          `json:"splits,omitempty"`
	SplitAmount         uint64                       `json:"splitamount,omitempty"`
	MaxOutPPM           uint64                       `json:"maxoutppm,omitempty"`
	DepleteUpToPercent  float64                      `json:"depleteuptopercent,omitempty"`
	DepleteUpToAmount   uint64                       `json:"depleteuptoamount,omitempty"`
	Attempts            int                          `json:"attempts,omitempty"`
	MaxHops             int                          `json:"maxhops,omitempty"`
	TotalAttempts       int                          `json:"-"`
	Node                *node.Node                   `json:"-"`
	InChannel           *graph.Channel               `json:"-"`
	Candidates          *deque.Deque[*graph.Channel] `json:"-"`
	AmountRebalanced    uint64                       `json:"-"`
	InFlightAmount      uint64                       `json:"-"`
	AmountLock          *sync.Cond                   `json:"-"`
	QueueLock           *sync.Cond                   `json:"-"`
	RebalanceResultChan chan *rebalance2.Result      `json:"-"`
}

func (r *RebalanceParallel) Name() string {
	return "circular-parallel"
}

func (r *RebalanceParallel) New() interface{} {
	return &RebalanceParallel{}
}

func (r *RebalanceParallel) Call() (jrpc2.Result, error) {
	r.Node = node.GetNode()
	if r.InScid == "" {
		return nil, util.ErrNoRequiredParameter
	}
	r.AmountLock = sync.NewCond(&sync.Mutex{})
	r.QueueLock = sync.NewCond(&sync.Mutex{})
	r.TotalAttempts = 0
	r.RebalanceResultChan = make(chan *rebalance2.Result)

	r.setDefaults()

	r.Node.Logf(glightning.Debug, "%+v", r)

	if err := r.validateParameters(); err != nil {
		return nil, err
	}

	incomingChannel, err := r.Node.GetIncomingChannelFromScid(r.InScid)
	if err != nil {
		return nil, err
	}
	r.InChannel = incomingChannel

	err = r.FindCandidates(r.InChannel.Source)
	if err != nil {
		return nil, err
	}

	r.FireCandidates()
	return r.WaitForResult()
}

func (r *RebalanceParallel) FireCandidates() {
	r.Node.Logln(glightning.Debug, "Firing candidates")
	r.AmountLock.L.Lock()
	carryOn := r.AmountRebalanced+r.InFlightAmount < r.Amount
	splitsInFlight := int(r.InFlightAmount / r.SplitAmount)
	r.AmountLock.L.Unlock()

	r.Node.Logln(glightning.Debug, "AmountRebalanced: ", r.AmountRebalanced, ", InFlightAmount: ", r.InFlightAmount, ", Total Amount:", r.Amount)
	r.Node.Logln(glightning.Debug, "Carry on: ", carryOn, ", Splits in flight: ", splitsInFlight)
	for carryOn && splitsInFlight < r.Splits {
		candidate, err := r.GetNextCandidate()
		if err != nil {
			// no candidate left
			r.Node.Logln(glightning.Debug, err)
			break
		}
		r.Fire(candidate)

		r.AmountLock.L.Lock()
		carryOn = r.AmountRebalanced+r.InFlightAmount < r.Amount
		splitsInFlight = int(r.InFlightAmount / r.SplitAmount)
		r.AmountLock.L.Unlock()
		r.Node.Logln(glightning.Debug, "Carry on: ", carryOn, ", Splits in flight: ", splitsInFlight)
	}
}

func (r *RebalanceParallel) WaitForResult() (jrpc2.Result, error) {
	start := time.Now()

	for r.InFlightAmount > 0 {
		r.Node.Logln(glightning.Debug, "Waiting for result, InFlightAmount:", r.InFlightAmount)
		rebalanceResult := <-r.RebalanceResultChan

		if rebalanceResult.Status == "status" {
			r.Node.Logf(glightning.Info, "Successful rebalance: %+v", rebalanceResult)
		} else {
			r.Node.Logf(glightning.Debug, "Failed rebalance: %+v", rebalanceResult)
		}

		// update inflight and rebalanced amount
		r.UpdateAmounts(rebalanceResult)

		// if we had a success, we put the candidate back in front of the queue
		if rebalanceResult.Status == "success" {
			r.EnqueueCandidate(rebalanceResult)
		}

		// now that we had a result, we can fire more candidates
		r.FireCandidates()
	}
	elapsed := time.Since(start)
	r.Node.Logf(glightning.Info, "Finished in %s", elapsed)
	return NewResult(r), nil
}

func (r *RebalanceParallel) Fire(candidate *graph.Channel) {
	r.Node.Logln(glightning.Debug, "Firing candidate:", candidate.ShortChannelId)
	r.TotalAttempts++
	rebalance := rebalance2.NewRebalance(candidate, r.InChannel, r.SplitAmount, r.MaxPPM, r.Attempts, r.MaxHops)

	r.AmountLock.L.Lock()
	r.InFlightAmount += rebalance.Amount
	r.AmountLock.L.Unlock()

	go func() {
		result, _ := rebalance.Run()
		r.RebalanceResultChan <- result
	}()
}

func (r *RebalanceParallel) UpdateAmounts(result *rebalance2.Result) {
	r.AmountLock.L.Lock()
	r.InFlightAmount -= r.SplitAmount
	if result.Status == "success" {
		r.AmountRebalanced += r.SplitAmount

		// not really a good way to do it, but we need to do this to make sure we don't
		// overshoot the Deplete amount. This is necessary because otherwise the
		// spendable balance would only be updated on refreshPeers.
		scid := result.Route.Hops[0].ShortChannelId
		r.Node.UpdateChannelBalance(result.Out, scid, result.Amount)
	}
	r.AmountLock.L.Unlock()
}
