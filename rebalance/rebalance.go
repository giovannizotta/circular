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
	"strconv"
	"time"
)

const (
	NORMAL           = "CHANNELD_NORMAL"
	DEFAULT_AMOUNT   = 200000000
	DEFAULT_PPM      = 100
	DEFAULT_ATTEMPTS = 10
)

type Rebalance struct {
	Out      string     `json:"out"`
	In       string     `json:"in"`
	Amount   uint64     `json:"amount,omitempty"`
	MaxPPM   uint64     `json:"maxppm,omitempty"`
	Attempts int        `json:"attempts,omitempty"`
	Node     *node.Node `json:"-"`
}

func (r *Rebalance) Name() string {
	return "circular"
}

func (r *Rebalance) New() interface{} {
	return &Rebalance{}
}

func (r *Rebalance) Call() (jrpc2.Result, error) {
	r.Node = node.GetNode()
	//convert to msatoshi
	r.Amount *= 1000

	log.Println("circular rebalance called")
	log.Println("self: ", r.Node.Id)
	log.Println("in:", r.In)
	log.Println("out:", r.Out)
	log.Println("amount:", r.Amount, "maxppm:", r.MaxPPM)
	log.Println("attempts:", r.Attempts)
	if err := r.validateParameters(); err != nil {
		return nil, err
	}

	for i := 0; i < r.Attempts; i++ {
		log.Println("===================== ATTEMPT", i+1, "=====================")
		result, ok := r.run()
		if ok {
			result += " after " + strconv.Itoa(i) + " attempts"
			log.Println(result)
			return NewResult(result), nil
		}
		if result != "TEMPORARY_FAILURE" {
			return NewResult(result), nil
		}
	}

	return NewResult("rebalance failed after " + strconv.Itoa(r.Attempts+1) + " attempts"), nil
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
			route.FeePPM(), r.MaxPPM))
	}

	return route, nil
}

func (r *Rebalance) tryRoute() (*graph.Route, error) {
	paymentSecret, err := r.Node.GeneratePreimageHashPair()
	if err != nil {
		return nil, err
	}

	route, err := r.getRoute()
	if err != nil {
		return nil, err
	}

	log.Println(route)

	_, err = r.Node.SendPay(route, paymentSecret)
	if err != nil {
		return nil, errors.New("TEMPORARY_FAILURE")
	}

	return route, nil
}

func (r *Rebalance) run() (string, bool) {
	defer util.TimeTrack(time.Now(), "rebalance.run")

	route, err := r.tryRoute()
	if err != nil {
		return err.Error(), false
	}

	return fmt.Sprintf(""+
		"successfully rebalanced %d sats "+
		"from %s to %s at %d ppm. Total fees paid: %.3f sats",
		r.Amount/1000, r.Out[:8], r.In[:8], route.FeePPM(), float64(route.Fee())/1000), true
}
