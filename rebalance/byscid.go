package rebalance

import (
	"circular/node"
	"circular/util"
	"github.com/elementsproject/glightning/jrpc2"
)

type RebalanceByScid struct {
	OutScid  string     `json:"outscid"`
	InScid   string     `json:"inscid"`
	Amount   uint64     `json:"amount,omitempty"`
	MaxPPM   uint64     `json:"maxppm,omitempty"`
	Attempts int        `json:"attempts,omitempty"`
	Node     *node.Node `json:"-"`
}

func (r *RebalanceByScid) Name() string {
	return "circular"
}

func (r *RebalanceByScid) New() interface{} {
	return &RebalanceByScid{}
}

func (r *RebalanceByScid) Call() (jrpc2.Result, error) {
	r.Node = node.GetNode()
	if r.InScid == "" || r.OutScid == "" {
		return nil, util.ErrNoRequiredParameter
	}

	outPeer, err := r.Node.GetChannelPeerFromScid(r.OutScid)
	if err != nil {
		return nil, err
	}
	inPeer, err := r.Node.GetChannelPeerFromScid(r.InScid)
	if err != nil {
		return nil, err
	}

	outgoingChannelId := r.OutScid + "/" + util.GetDirection(r.Node.Id, outPeer.Id)
	incomingChannelId := r.InScid + "/" + util.GetDirection(inPeer.Id, r.Node.Id)

	outgoingChannel := r.Node.Graph.Channels[outgoingChannelId]
	incomingChannel := r.Node.Graph.Channels[incomingChannelId]

	rebalance := NewRebalance(outgoingChannel, incomingChannel, r.Amount, r.MaxPPM, r.Attempts)

	err = rebalance.Setup()
	if err != nil {
		return nil, err
	}

	return rebalance.Run()
}
