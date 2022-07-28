package parallel

import (
	"circular/node"
	"circular/util"
	"github.com/elementsproject/glightning/jrpc2"
)

type RebalanceParallel struct {
	InScid             string     `json:"inscid"`
	Amount             uint64     `json:"amount,omitempty"`
	MaxPPM             uint64     `json:"maxppm,omitempty"`
	Splits             int        `json:"splits,omitempty"`
	SplitAmount        uint64     `json:"splitamount,omitempty"`
	MaxOutPPM          uint64     `json:"maxoutppm,omitempty"`
	DepleteUpToPercent float64    `json:"depleteuptopercent,omitempty"`
	DepleteUpToAmount  uint64     `json:"depleteuptoamount,omitempty"`
	Attempts           int        `json:"attempts,omitempty"`
	MaxHops            int        `json:"maxhops,omitempty"`
	Node               *node.Node `json:"-"`
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
	r.setDefaults()

	if err := r.validateParameters(); err != nil {
		return nil, err
	}

	inPeer, err := r.Node.GetChannelPeerFromScid(r.InScid)
	if err != nil {
		return nil, err
	}

	incomingChannelId := r.InScid + "/" + util.GetDirection(r.Node.Id, inPeer.Id)
	if _, ok := r.Node.Graph.Channels[incomingChannelId]; !ok {
		return nil, util.ErrNoIncomingChannel
	}
	incomingChannel := r.Node.Graph.Channels[incomingChannelId]

}
