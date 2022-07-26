package rebalance

import (
	"circular/graph"
	"circular/node"
	"circular/util"
	"errors"
	"github.com/elementsproject/glightning/glightning"
	"github.com/elementsproject/glightning/jrpc2"
)

type RebalanceByNode struct {
	OutNode  string     `json:"outnode"`
	InNode   string     `json:"innode"`
	Amount   uint64     `json:"amount,omitempty"`
	MaxPPM   uint64     `json:"maxppm,omitempty"`
	Attempts int        `json:"attempts,omitempty"`
	MaxHops  int        `json:"maxhops,omitempty"`
	Node     *node.Node `json:"-"`
}

func (r *RebalanceByNode) Name() string {
	return "circular-node"
}

func (r *RebalanceByNode) New() interface{} {
	return &RebalanceByNode{}
}

func (r *RebalanceByNode) getBestOutgoingChannel() (*graph.Channel, error) {
	bestScid := r.Node.GetBestPeerChannel(r.OutNode, func(channel *glightning.PeerChannel) uint64 {
		return channel.SpendableMilliSatoshi
	}).ShortChannelId
	channelId := bestScid + "/" + util.GetDirection(r.Node.Id, r.OutNode)
	if channelId, ok := r.Node.Graph.Channels[channelId]; ok {
		return channelId, nil
	}
	return nil, util.ErrNoOutgoingChannel
}

func (r *RebalanceByNode) getBestIncomingChannel() (*graph.Channel, error) {
	bestScid := r.Node.GetBestPeerChannel(r.InNode, func(channel *glightning.PeerChannel) uint64 {
		return channel.ReceivableMilliSatoshi
	}).ShortChannelId
	channelId := bestScid + "/" + util.GetDirection(r.InNode, r.Node.Id)
	if channelId, ok := r.Node.Graph.Channels[channelId]; ok {
		return channelId, nil
	}
	return nil, util.ErrNoIncomingChannel
}

func (r *RebalanceByNode) Call() (jrpc2.Result, error) {
	r.Node = node.GetNode()
	if r.InNode == "" || r.OutNode == "" {
		return nil, util.ErrNoRequiredParameter
	}

	r.Node.RefreshPeer(r.OutNode)
	r.Node.RefreshPeer(r.InNode)

	err := r.validatePeers()
	if err != nil {
		return nil, err
	}

	// get channels from the nodes
	outgoingChannel, err := r.getBestOutgoingChannel()
	if err != nil {
		return nil, err
	}
	incomingChannel, err := r.getBestIncomingChannel()
	if err != nil {
		return nil, err
	}

	rebalance := NewRebalance(outgoingChannel, incomingChannel, r.Amount, r.MaxPPM, r.Attempts, r.MaxHops)

	err = rebalance.Setup()
	if err != nil {
		return nil, err
	}

	return rebalance.Run()
}

func (r *RebalanceByNode) validatePeers() error {
	if len(r.Node.Peers) == 0 {
		return util.ErrNoPeers
	}
	//validate that the nodes are not self
	if r.InNode == r.Node.Id || r.OutNode == r.Node.Id {
		return util.ErrSelfNode
	}
	//validate that the nodes are not the same
	if r.InNode == r.OutNode {
		return errors.New("incoming and outgoing nodes are the same")
	}

	//validate that the nodes are actually peers
	if _, ok := r.Node.Peers[r.InNode]; !ok {
		return util.ErrNoPeer
	}
	if _, ok := r.Node.Peers[r.OutNode]; !ok {
		return util.ErrNoPeer
	}
	return nil
}
