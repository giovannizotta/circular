package rebalance

import (
	"circular/graph"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"log"
	"time"
)

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

func (r *Rebalance) getRoute(maxHops int) (*graph.Route, error) {
	defer util.TimeTrack(time.Now(), "rebalance.getRoute")
	exclude := make(map[string]bool)
	exclude[r.Node.Id] = true

	route, err := r.Node.Graph.GetRoute(r.Out, r.In, r.Amount, exclude, maxHops)
	if err != nil {
		return nil, err
	}

	r.prependNode(route)
	r.appendNode(route)

	if route.FeePPM() > r.MaxPPM {
		log.Println("best route found was: ", route)
		return nil, NewRouteTooExpensiveError(route.FeePPM(), r.MaxPPM)
	}

	return route, nil
}

func (r *Rebalance) tryRoute(maxHops int) (*graph.Route, error) {
	paymentSecret, err := r.Node.GeneratePreimageHashPair()
	if err != nil {
		return nil, err
	}

	route, err := r.getRoute(maxHops)
	if err != nil {
		return nil, err
	}

	log.Println(route)

	_, err = r.Node.SendPay(route, paymentSecret)
	if err != nil {
		// TODO: meh, refactor
		if err.Error() == ErrSendPayTimeout.Error() {
			return nil, ErrSendPayTimeout
		}
		return nil, ErrTemporaryFailure
	}

	return route, nil
}
