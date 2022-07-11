package rebalance

import (
	"circular/graph"
	"circular/node"
	"circular/util"
	"errors"
	"fmt"
	"github.com/elementsproject/glightning/glightning"
	"github.com/elementsproject/glightning/jrpc2"
	"log"
	"time"
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
	Node   *node.Node `json:"-"`
}

func (r *Rebalance) Name() string {
	return "rebalance"
}

func (r *Rebalance) New() interface{} {
	return &Rebalance{}
}

func (r *Rebalance) Call() (jrpc2.Result, error) {
	r.Node = node.GetNode()
	log.Println("rebalance called")
	log.Println("self: ", r.Node.Id)
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

func (r *Rebalance) prependNode(route *graph.Route) {
	bestScid := r.Node.GetBestPeerChannel(r.Out, func(channel *glightning.PeerChannel) uint64 {
		return channel.SpendableMilliSatoshi
	}).ShortChannelId
	channelId := bestScid + "/" + util.GetDirection(r.Node.Id, r.Out)
	channel := r.Node.Graph.Channels[channelId]
	route.Prepend(channel)
}

func (r *Rebalance) appendNode(route *graph.Route) {
	bestScid := r.Node.GetBestPeerChannel(r.In, func(channel *glightning.PeerChannel) uint64 {
		return channel.ReceivableMilliSatoshi
	}).ShortChannelId
	channelId := bestScid + "/" + util.GetDirection(r.In, r.Node.Id)
	channel := r.Node.Graph.Channels[channelId]
	route.Append(channel)
}

func (r *Rebalance) getRoute() (*graph.Route, error) {
	defer util.TimeTrack(time.Now(), "rebalance.getRoute")
	exclude := make(map[string]bool)
	exclude[r.Node.Id] = true

	route, err := r.Node.Graph.GetRoute(r.Out, r.In, r.Amount, exclude)
	if err != nil {
		return nil, err
	}

	r.prependNode(route)
	r.appendNode(route)

	if route.FeePPM() > r.MaxPPM {
		return nil, errors.New(fmt.Sprintf("route too expensive. "+
			"Cheapest route found was %d ppm, but max_ppm is %d",
			route.FeePPM()/1000, r.MaxPPM/1000))
	}

	return route, nil
}

func (r *Rebalance) tryRoute() (*graph.Route, error) {
	defer util.TimeTrack(time.Now(), "rebalance.tryRoute")
	log.Println("generating preimage/hash pair")
	paymentSecret, err := r.Node.GeneratePreimageHashPair()
	if err != nil {
		return nil, err
	}

	route, err := r.getRoute()
	if err != nil {
		return nil, err
	}

	log.Println("Trying route with ppm ", route.FeePPM()/1000)
	log.Println("Hops: ", len(route.Hops))
	for _, hop := range route.Hops {
		log.Println("Hop: ", hop.Destination)
	}

	_, err = r.Node.SendPay(route, paymentSecret)
	if err != nil {
		return nil, err
	}

	return route, nil
}

func (r *Rebalance) run() (string, error) {
	defer util.TimeTrack(time.Now(), "rebalance.run")

	route, err := r.tryRoute()
	if err != nil {
		return "", err
	}

	// TODO: after successful rebalance, clean PreimageStore and refresh channel balances
	return fmt.Sprintf(""+
		"successfully rebalanced %d sats "+
		"from %s to %s at %d ppm. Total fees paid: %.3f sats",
		r.Amount/1000, r.Out[:8], r.In[:8], route.FeePPM()/1000, float64(route.Fee())/1000), nil
}
