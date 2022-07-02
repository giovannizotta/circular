package main

import (
	"errors"
	"fmt"
	"github.com/elementsproject/glightning/glightning"
	"github.com/elementsproject/glightning/jrpc2"
	"log"
)

const (
	NORMAL = "CHANNELD_NORMAL"
)

type Rebalance struct {
	In     string `json:"required"`
	Out    string `json:"required"`
	Amount uint64 `json:"required"`
}

func (r *Rebalance) Name() string {
	return "rebalance"
}

func (r *Rebalance) New() interface{} {
	return &Rebalance{}
}

func getPeerChannels(id string) []*glightning.PeerChannel {
	peer, err := lightning.GetPeer(id)
	if err != nil {
		log.Fatalln(err)
	}
	return peer.Channels
}

func getBestChannel(peer string, metric func(channel *glightning.PeerChannel) uint64) *glightning.PeerChannel {
	channels := getPeerChannels(peer)
	best := channels[0]
	for _, channel := range channels {
		if metric(channel) > metric(best) {
			best = channel
		}
	}
	return best
}

func (r *Rebalance) validateRebalancePeerParameters() error {
	//validate that the nodes are not self
	if r.In == self.Id || r.Out == self.Id {
		return errors.New("one of the nodes is self")
	}
	//validate that the nodes are not the same
	if r.In == r.Out {
		return errors.New("incoming and outgoing nodes are the same")
	}
	//validate that the r.In is a neighbor of self
	if _, ok := graph.Nodes[self.Id][r.In]; !ok {
		return errors.New("incoming node is not a peer")
	}
	//validate r.Out is in graph.Nodes[self]
	if _, ok := graph.Nodes[self.Id][r.Out]; !ok {
		return errors.New("outgoing node is not a peer")
	}
	return nil
}

func (r *Rebalance) validateRebalanceLiquidityParameters() error {
	inChannel := getBestChannel(r.In, func(channel *glightning.PeerChannel) uint64 {
		return channel.ReceivableMilliSatoshi
	})
	outChannel := getBestChannel(r.Out, func(channel *glightning.PeerChannel) uint64 {
		return channel.SpendableMilliSatoshi
	})
	log.Printf("inChannel: %v\n", inChannel)
	log.Printf("outChannel: %v\n", outChannel)
	//validate that the channels are in normal state
	if inChannel.State != NORMAL {
		return errors.New("incoming channel is not in normal state")
	}
	if outChannel.State != NORMAL {
		return errors.New("outgoing channel is not in normal state")
	}
	//validate that the amount is less than the liquidity of the channels
	if (inChannel.ReceivableMilliSatoshi / 1000) < r.Amount {
		return errors.New("incoming channel has insufficient remote balance")
	}
	if (outChannel.SpendableMilliSatoshi / 1000) < r.Amount {
		return errors.New("outgoing channel has insufficient local balance")
	}
	return nil
}

func (r *Rebalance) validateRebalanceParameters() error {
	if r.In == "" || r.Out == "" || r.Amount <= 0 {
		return errors.New("missing required parameter")
	}
	err := r.validateRebalancePeerParameters()
	if err != nil {
		return err
	}

	err = r.validateRebalanceLiquidityParameters()
	if err != nil {
		return err
	}
	return nil
}

type RebalanceResult struct {
	Result     string `json:"rebalance"`
	FormatHint string `json:"format-hint,omitempty"`
}

func NewRebalanceResult(result string) *RebalanceResult {
	return &RebalanceResult{
		Result:     result,
		FormatHint: glightning.FormatSimple,
	}
}

func (r *Rebalance) Call() (jrpc2.Result, error) {
	err := r.validateRebalanceParameters()
	if err != nil {
		return nil, err
	}
	return NewRebalanceResult(
		fmt.Sprintf("Rebalancing:\nin: %s\nout: %s\namount: %d\n", r.In, r.Out, r.Amount)), nil
}
