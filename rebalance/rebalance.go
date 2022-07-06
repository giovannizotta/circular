package rebalance

import (
	"circular/graph"
	"circular/node"
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
	In               string           `json:"in"`
	Out              string           `json:"out"`
	Amount           uint64           `json:"amount"`
	MaxPPM           uint64           `json:"max_ppm"`
	PreimageHashPair PreimageHashPair `json:"preimage,omitempty"`
	Self             *node.Self       `json:"self,omitempty"`
}

func (r *Rebalance) Name() string {
	return "rebalance"
}

func (r *Rebalance) New() interface{} {
	return &Rebalance{}
}

func (r *Rebalance) Call() (jrpc2.Result, error) {
	log.Println("rebalance called")
	r.Self = node.GetSelf()
	if err := r.validateParameters(); err != nil {
		return nil, err
	}

	log.Printf("parameters validated, running rebalance\n")
	//convert to msatoshi
	r.Amount *= 1000
	r.MaxPPM *= 1000

	result, err := r.run()
	if err != nil {
		return nil, err
	}
	return NewResult(result), nil
}

func (r *Rebalance) validatePeerParameters() error {
	//validate that the nodes are not self
	if r.In == r.Self.Id || r.Out == r.Self.Id {
		return errors.New("one of the nodes is self")
	}
	//validate that the nodes are not the same
	if r.In == r.Out {
		return errors.New("incoming and outgoing nodes are the same")
	}
	if len(r.Self.Peers) == 0 {
		return errors.New("no peers yet")
	}
	//validate that In and Out are peers
	if !r.Self.HasPeer(r.In) {
		return errors.New("incoming value is not a peer")
	}
	if !r.Self.HasPeer(r.Out) {
		return errors.New("outgoing value is not a peer")
	}
	return nil
}

func (r *Rebalance) validateLiquidityParameters() error {
	inChannel := r.Self.GetBestPeerChannel(r.In, func(channel *glightning.PeerChannel) uint64 {
		return channel.ReceivableMilliSatoshi
	})
	outChannel := r.Self.GetBestPeerChannel(r.Out, func(channel *glightning.PeerChannel) uint64 {
		return channel.SpendableMilliSatoshi
	})
	//validate that the channels are in normal state
	if inChannel.State != NORMAL {
		return errors.New("incoming channel is not in normal state")
	}
	if outChannel.State != NORMAL {
		return errors.New("outgoing channel is not in normal state")
	}
	//validate that the amount is less than the liquidity of the channels
	if (inChannel.ReceivableMilliSatoshi) < r.Amount {
		return errors.New("incoming channel has insufficient remote balance")
	}
	if (outChannel.SpendableMilliSatoshi) < r.Amount {
		return errors.New("outgoing channel has insufficient local balance")
	}
	return nil
}

func (r *Rebalance) validateParameters() error {
	if r.In == "" || r.Out == "" || r.Amount <= 0 || r.MaxPPM < 0 {
		return errors.New("missing required parameter")
	}
	err := r.validatePeerParameters()
	if err != nil {
		return err
	}

	err = r.validateLiquidityParameters()
	if err != nil {
		return err
	}

	return nil
}

func (r *Rebalance) getRoute() (*graph.Route, error) {
	route, err := graph.NewRoute(r.Self.Graph, r.In, r.Out, r.Amount, []string{r.Self.Id})
	if err != nil {
		return nil, err
	}

	// prepend self to the route
	bestOutgoingScid := r.Self.GetBestPeerChannel(r.Out, func(channel *glightning.PeerChannel) uint64 {
		return channel.ReceivableMilliSatoshi
	}).ShortChannelId
	route.PrependHop(r.Self.Id, bestOutgoingScid)

	// append self to the route
	bestIncomingScid := r.Self.GetBestPeerChannel(r.In, func(channel *glightning.PeerChannel) uint64 {
		return channel.SpendableMilliSatoshi
	}).ShortChannelId
	route.AppendHop(r.Self.Id, bestIncomingScid)

	for i, hop := range route.Hops {
		log.Printf("hop %d: %+v\n", i, hop)
	}

	if route.FeePPM > r.MaxPPM {
		return nil, errors.New(fmt.Sprintf("graph too expensive. "+
			"Cheapest graph found was %d ppm, but max_ppm is %d",
			route.FeePPM/1000, r.MaxPPM/1000))
	}

	return route, nil
}

func (r *Rebalance) run() (string, error) {
	//TODO: save the preimage hash pair in the database
	log.Println("generating preimage/hash pair")
	r.PreimageHashPair = *NewPreimageHashPair()
	r.Self.OngoingRebalances[r.PreimageHashPair.Hash] = r.PreimageHashPair.Preimage

	log.Println("searching for a route")
	route, err := r.getRoute()
	if err != nil {
		return "", err
	}

	log.Println("trying to send payment to route")
	_, err = r.Self.SendPay(route, r.PreimageHashPair.Hash)
	if err != nil {
		return "", err
	}

	// TODO: after successful rebalance, clean DB and refresh channel balances
	return fmt.Sprintf("rebalance successful at %d ppm\n", route.FeePPM/1000), nil
}
