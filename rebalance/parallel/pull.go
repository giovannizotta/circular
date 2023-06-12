package parallel

import (
	"circular/graph"
	rebalance2 "circular/rebalance"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"github.com/elementsproject/glightning/jrpc2"
)

const (
	DEFAULT_MAX_OUT_PPM           = 50
	DEFAULT_DEPLETE_UP_TO_PERCENT = 0.2
	DEFAULT_DEPLETE_UP_TO_AMOUNT  = 1000000
)

type RebalancePull struct {
	InScid             string   `json:"inscid"`
	OutList            []string `json:"outlist,omitempty"`
	MaxOutPPM          uint64   `json:"maxoutppm,omitempty"`
	Amount             uint64   `json:"amount,omitempty"`
	MaxPPM             uint64   `json:"maxppm,omitempty"`
	Splits             int      `json:"splits,omitempty"`
	SplitAmount        uint64   `json:"splitamount,omitempty"`
	DepleteUpToPercent float64  `json:"depleteuptopercent,omitempty"`
	DepleteUpToAmount  uint64   `json:"depleteuptoamount,omitempty"`
	Attempts           int      `json:"attempts,omitempty"`
	MaxHops            int      `json:"maxhops,omitempty"`
	AbstractRebalance
}

func (r *RebalancePull) Name() string {
	return "circular-pull"
}

func (r *RebalancePull) New() interface{} {
	return &RebalancePull{}
}

func (r *RebalancePull) Call() (jrpc2.Result, error) {
	r.AbstractRebalance.RebalanceMethods = r
	if r.InScid == "" {
		return nil, util.ErrNoRequiredParameter
	}
	r.Init(r.Amount, r.MaxPPM, r.SplitAmount, r.Splits, r.Attempts, r.MaxHops)

	r.CandidatesList = r.OutList
	if r.CandidatesList != nil {
		r.Node.Logln(glightning.Info, "Using outlist:", r.CandidatesList)
		// if an outlist was supplied, ignore maxoutppm. To do this we put it to "infinity"
		r.MaxOutPPM = 1 << 63
	}

	r.setDefaults()

	if err := r.validateParameters(); err != nil {
		return nil, err
	}
	r.Node.Logf(glightning.Debug, "RebalancePull parameters validated: %+v", r)

	incomingChannel, err := r.Node.GetIncomingChannelFromScid(r.InScid)
	if err != nil {
		return nil, err
	}
	r.TargetChannel = incomingChannel

	if err = r.FindCandidates(r.TargetChannel.Source); err != nil {
		return nil, err
	}

	r.FireCandidates()
	return r.WaitForResult()
}

func (r *RebalancePull) IsGoodCandidate(peerChannel *glightning.PeerChannel) bool {
	// we need to get the outgoing channel from the peer to compute outgoing PPM and check it's below the maxoutppm
	outgoingChannel, err := r.Node.GetOutgoingChannelFromScid(peerChannel.ShortChannelId)
	if err != nil {
		r.Node.Logln(glightning.Unusual, err)
		return false
	}

	return outgoingChannel.ComputeFeePPM(r.splitAmount) < r.MaxOutPPM
}

// Check that the channel is not under the deplete threshold and connection is active
func (r *RebalancePull) CanUseChannel(channel *glightning.PeerChannel) error {
	depleteAmount := util.Min(r.DepleteUpToAmount,
		uint64(float64(channel.TotalMsat.MSat())*r.DepleteUpToPercent))
	r.Node.Logln(glightning.Debug, "depleteAmount:", depleteAmount)
	if channel.ToUsMsat.MSat() < depleteAmount {
		return util.ErrChannelDepleted
	}

	if channel.State != rebalance2.NORMAL {
		return util.ErrChannelNotInNormalState
	}

	if r.Node.IsPeerConnected(channel) == false {
		return util.ErrOutgoingPeerDisconnected
	}

	return nil
}

func (r *RebalancePull) Fire(candidate *graph.Channel) {
	r.Node.Logln(glightning.Debug, "Firing candidate: ", candidate.ShortChannelId, " for attempts: ", r.attempts)
	rebalance := rebalance2.NewRebalance(candidate, r.TargetChannel, r.splitAmount, r.maxPPM, r.attempts, r.maxHops)

	go func() {
		r.RebalanceResultChan <- rebalance.Run()
	}()
}

func (r *RebalancePull) setDefaults() {
	if r.DepleteUpToPercent <= 0 {
		r.DepleteUpToPercent = DEFAULT_DEPLETE_UP_TO_PERCENT
	}
	if r.DepleteUpToAmount == 0 {
		r.DepleteUpToAmount = DEFAULT_DEPLETE_UP_TO_AMOUNT
	}
	if r.MaxOutPPM == 0 {
		r.MaxOutPPM = DEFAULT_MAX_OUT_PPM
	}

	// convert to msat
	r.DepleteUpToAmount *= 1000
}

func (r *RebalancePull) validateParameters() error {
	if err := r.validateGenericParameters(); err != nil {
		return err
	}
	if r.DepleteUpToPercent > 1 || r.DepleteUpToPercent < 0 {
		return util.ErrDepleteUpToPercentInvalid
	}
	return nil
}

func (r *RebalancePull) GetCandidateDirection(id string) string {
	return util.GetDirection(r.Node.Id, id)
}

// EnqueueCandidate puts a candidate at the front of the queue
func (r *RebalancePull) EnqueueCandidate(result *rebalance2.Result) {
	scid := result.Route.Hops[0].ShortChannelId
	candidate, err := r.Node.GetOutgoingChannelFromScid(scid)
	if err != nil {
		r.Node.Logln(glightning.Unusual, err)
		return
	}

	r.QueueLock.Lock()
	r.Candidates.PushFront(candidate)
	r.QueueLock.Unlock()
}

func (r *RebalancePull) AddSuccess(result *rebalance2.Result) {
	r.Node.Graph.LockAliases()
	defer r.Node.Graph.UnlockAliases()
	alias := "unknown"
	if a, ok := r.Node.Graph.Aliases[result.Out]; ok {
		alias = a
	}
	r.AddSuccessGeneric(alias, result.PPM, result.Amount)
}
