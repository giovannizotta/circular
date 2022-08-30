package parallel

import (
	"circular/graph"
	rebalance2 "circular/rebalance"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"github.com/elementsproject/glightning/jrpc2"
)

const (
	DEFAULT_MIN_OUT_PPM        = 50
	DEFAULT_FILL_UP_TO_PERCENT = 0.8
	DEFAULT_FILL_UP_TO_AMOUNT  = 10000000
)

type RebalancePush struct {
	OutScid         string   `json:"outscid"`
	InList          []string `json:"inlist,omitempty"`
	MinOutPPM       uint64   `json:"minoutppm,omitempty"`
	Amount          uint64   `json:"amount,omitempty"`
	MaxPPM          uint64   `json:"maxppm,omitempty"`
	Splits          int      `json:"splits,omitempty"`
	SplitAmount     uint64   `json:"splitamount,omitempty"`
	Attempts        int      `json:"attempts,omitempty"`
	MaxHops         int      `json:"maxhops,omitempty"`
	FillUpToPercent float64  `json:"filluptopercent,omitempty"`
	FillUpToAmount  uint64   `json:"filluptoamount,omitempty"`
	AbstractRebalance
}

func (r *RebalancePush) Name() string {
	return "circular-push"
}

func (r *RebalancePush) New() interface{} {
	return &RebalancePush{}
}

func (r *RebalancePush) Call() (jrpc2.Result, error) {
	r.AbstractRebalance.RebalanceMethods = r
	if r.OutScid == "" {
		return nil, util.ErrNoRequiredParameter
	}
	r.Init(r.Amount, r.MaxPPM, r.SplitAmount, r.Splits, r.Attempts, r.MaxHops)

	r.CandidatesList = r.InList
	if r.CandidatesList != nil {
		r.Node.Logln(glightning.Info, "Using inlist:", r.CandidatesList)
		// if an inlist was supplied, ignore minoutppm. To do this we put it to zero
		r.MinOutPPM = 0
	}

	r.setDefaults()

	if err := r.validateParameters(); err != nil {
		return nil, err
	}

	outgoingChannel, err := r.Node.GetOutgoingChannelFromScid(r.OutScid)
	if err != nil {
		return nil, err
	}
	r.TargetChannel = outgoingChannel

	if err = r.FindCandidates(r.TargetChannel.Destination); err != nil {
		return nil, err
	}

	r.FireCandidates()
	return r.WaitForResult()
}

func (r *RebalancePush) IsGoodCandidate(peerChannel *glightning.PeerChannel) bool {
	// first of all, if the peer charges a higher fee than maxppm towards us, there's no point in trying to use it
	incomingChannel, err := r.Node.GetIncomingChannelFromScid(peerChannel.ShortChannelId)
	if err != nil {
		r.Node.Logln(glightning.Unusual, err)
		return false
	}
	if incomingChannel.ComputeFeePPM(r.splitAmount) > r.maxPPM {
		return false
	}

	// we need to get the outgoing channel from the peer to compute outgoing PPM and check it's above the minoutppm
	outgoingChannel, err := r.Node.GetOutgoingChannelFromScid(peerChannel.ShortChannelId)
	if err != nil {
		r.Node.Logln(glightning.Unusual, err)
		return false
	}

	return outgoingChannel.ComputeFeePPM(r.splitAmount) > r.MinOutPPM
}

func (r *RebalancePush) CanUseChannel(channel *glightning.PeerChannel) error {
	// check that the channel is not over the fill threshold
	fillAmount := util.Min(r.FillUpToAmount,
		uint64(float64(channel.MilliSatoshiTotal)*r.FillUpToPercent))
	r.Node.Logln(glightning.Debug, "fillAmount:", fillAmount)
	if (channel.MilliSatoshiTotal - channel.MilliSatoshiToUs) < fillAmount {
		return util.ErrChannelFilled
	}

	if channel.State != rebalance2.NORMAL {
		return util.ErrChannelNotInNormalState
	}

	if r.Node.IsPeerConnected(channel) == false {
		return util.ErrIncomingPeerDisconnected
	}

	return nil
}

func (r *RebalancePush) Fire(candidate *graph.Channel) {
	r.Node.Logln(glightning.Info, "Firing candidate: ", candidate.ShortChannelId)
	rebalance := rebalance2.NewRebalance(r.TargetChannel, candidate, r.splitAmount, r.maxPPM, r.attempts, r.maxHops)

	go func() {
		r.RebalanceResultChan <- rebalance.Run()
	}()
}

func (r *RebalancePush) validateParameters() error {
	if err := r.validateGenericParameters(); err != nil {
		return err
	}
	if r.FillUpToPercent > 1 || r.FillUpToPercent < 0 {
		return util.ErrDepleteUpToPercentInvalid
	}
	return nil
}

func (r *RebalancePush) setDefaults() {
	if r.MinOutPPM == 0 {
		r.MinOutPPM = DEFAULT_MIN_OUT_PPM
	}
	if r.FillUpToPercent <= 0 {
		r.FillUpToPercent = DEFAULT_FILL_UP_TO_PERCENT
	}
	if r.FillUpToAmount == 0 {
		r.FillUpToAmount = DEFAULT_FILL_UP_TO_AMOUNT
	}

	// convert to msat
	r.FillUpToAmount *= 1000
}

func (r *RebalancePush) GetCandidateDirection(id string) string {
	return util.GetDirection(id, r.Node.Id)
}

// EnqueueCandidate puts a candidate at the front of the queue
func (r *RebalancePush) EnqueueCandidate(result *rebalance2.Result) {
	scid := result.Route.Hops[len(result.Route.Hops)-1].ShortChannelId
	candidate, err := r.Node.GetIncomingChannelFromScid(scid)
	if err != nil {
		r.Node.Logln(glightning.Unusual, err)
		return
	}

	r.QueueLock.Lock()
	r.Candidates.PushFront(candidate)
	r.QueueLock.Unlock()
}
