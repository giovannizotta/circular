package parallel

import (
	"circular/rebalance"
	"fmt"
	"github.com/elementsproject/glightning/glightning"
	"github.com/elementsproject/glightning/jrpc2"
	"time"
)

// Success is a map from PPM to amount
// for each PPM, the amount of sats rebalanced at that ppm
type Success map[uint64]uint64

type Result struct {
	RebalanceTarget  uint64             `json:"rebalance_target"`
	RebalancedAmount uint64             `json:"rebalanced_amount"`
	Attempts         uint64             `json:"attempts"`
	Time             string             `json:"time"`
	Successes        map[string]Success `json:"successes"`
}

func NewResult(target uint64) *Result {
	return &Result{
		RebalanceTarget:  target / 1000,
		RebalancedAmount: 0,
		Successes:        make(map[string]Success),
	}
}

func (r *AbstractRebalance) AddSuccessGeneric(alias string, ppm, amount uint64) {
	r.Result.RebalancedAmount += amount
	if _, ok := r.Result.Successes[alias]; !ok {
		r.Result.Successes[alias] = make(map[uint64]uint64)
	}
	if _, ok := r.Result.Successes[alias][ppm]; !ok {
		r.Result.Successes[alias][ppm] = 0
	}
	r.Result.Successes[alias][ppm] += amount
}

func (r *AbstractRebalance) WaitForResult() (jrpc2.Result, error) {
	start := time.Now()
	r.Result = NewResult(r.amount)

	// while there's something inflight, wait for results
	for r.InFlightAmount > 0 {
		r.Node.Logln(glightning.Debug, "Waiting for result, InFlightAmount:", r.InFlightAmount)
		rebalanceResult := <-r.RebalanceResultChan

		r.TotalAttempts += rebalanceResult.Attempts

		if rebalanceResult.Status == "success" {
			r.Node.Logf(glightning.Info, "Successful rebalance: %+v", rebalanceResult)

			// update results data
			r.AddSuccess(rebalanceResult)

			// put the candidate back in front of the queue
			r.EnqueueCandidate(rebalanceResult)
		} else {
			r.Node.Logf(glightning.Debug, "Failed rebalance: %+v", rebalanceResult)
		}

		// update inflight and rebalanced amount
		r.UpdateAmounts(rebalanceResult)

		// now that we had a result, we can fire more candidates
		r.FireCandidates()
	}

	// rebalance is over
	r.Result.Attempts = r.TotalAttempts
	r.Result.Time = fmt.Sprintf("%.3fs", float64(time.Since(start).Milliseconds())/1000)
	return r.Result, nil
}

func (r *AbstractRebalance) UpdateAmounts(result *rebalance.Result) {
	r.AmountLock.Lock()
	defer r.AmountLock.Unlock()

	r.InFlightAmount -= r.splitAmount
	if result.Status == "success" {
		r.AmountRebalanced += r.splitAmount

		// not really a good way to do it, but we need to do this to make sure we don't
		// overshoot the Deplete/Fill amount. This is necessary because otherwise the
		// spendable balance would only be updated on refreshPeers.
		outScid := result.Route.Hops[0].ShortChannelId
		inScid := result.Route.Hops[len(result.Route.Hops)-1].ShortChannelId
		r.Node.UpdateChannelBalance(result.Out, result.In, outScid, inScid, result.Amount)
	}
}
