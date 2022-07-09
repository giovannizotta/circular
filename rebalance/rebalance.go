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
	NORMAL         = "CHANNELD_NORMAL"
	DEFAULT_AMOUNT = 200000
	DEFAULT_PPM    = 100
)

type Rebalance struct {
	In     string     `json:"in"`
	Out    string     `json:"out"`
	Amount uint64     `json:"amount,omitempty"`
	MaxPPM uint64     `json:"max_ppm,omitempty"`
	Self   *node.Node `json:"self,omit"`
}

func (r *Rebalance) Name() string {
	return "rebalance"
}

func (r *Rebalance) New() interface{} {
	return &Rebalance{}
}

func (r *Rebalance) Call() (jrpc2.Result, error) {
	r.Self = node.GetNode()
	log.Println("rebalance called")
	log.Println("self: ", r.Self.Id)
	log.Println("in:", r.In)
	log.Println("out:", r.Out)
	log.Println("amount:", r.Amount, "max_ppm:", r.MaxPPM)
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

func (r *Rebalance) insertSelfInRoute(route *graph.Route) {
	// prepend self to the route
	bestOutgoingScid := r.Self.GetBestPeerChannel(r.Out, func(channel *glightning.PeerChannel) uint64 {
		return channel.ReceivableMilliSatoshi
	}).ShortChannelId
	outgoingChannel := r.Self.Graph.Outbound[r.Self.Id][r.Out][bestOutgoingScid]
	route.Prepend(outgoingChannel)

	// append self to the route
	bestIncomingScid := r.Self.GetBestPeerChannel(r.In, func(channel *glightning.PeerChannel) uint64 {
		return channel.SpendableMilliSatoshi
	}).ShortChannelId
	incomingChannel := r.Self.Graph.Outbound[r.In][r.Self.Id][bestIncomingScid]
	route.Append(incomingChannel)
}

func (r *Rebalance) getRoute() (*graph.Route, error) {
	exclude := make(map[string]bool)
	exclude[r.Self.Id] = true

	route, err := r.Self.Graph.GetRoute(r.Out, r.In, r.Amount, exclude)
	if err != nil {
		return nil, err
	}

	r.insertSelfInRoute(route)

	if route.FeePPM() > r.MaxPPM {
		return nil, errors.New(fmt.Sprintf("route too expensive. "+
			"Cheapest route found was %d ppm, but max_ppm is %d",
			route.FeePPM()/1000, r.MaxPPM/1000))
	}

	return route, nil
}

func (r *Rebalance) run() (string, error) {
	log.Println("generating preimage/hash pair")
	paymentSecret, err := r.Self.GeneratePreimageHashPair()
	if err != nil {
		return "", err
	}

	log.Println("searching for a route")
	route, err := r.getRoute()
	if err != nil {
		return "", err
	}

	log.Println("trying to send payment to route")
	_, err = r.Self.SendPay(route, paymentSecret)
	if err != nil {
		return "", err
	}

	// TODO: after successful rebalance, clean PreimageStore and refresh channel balances
	return fmt.Sprintf(""+
		"successfully rebalanced %d sats "+
		"from %s to %s at %d ppm. Total fees paid: %.3f sats",
		r.Amount/1000, r.Out[:8], r.In[:8], route.FeePPM()/1000, float64(route.Fee())/1000), nil
}
