package rebalance

import (
	"circular/graph"
	"circular/util"
	"log"
	"time"
)

func (r *Rebalance) getRoute(maxHops int) (*graph.Route, error) {
	defer util.TimeTrack(time.Now(), "rebalance.getRoute")
	exclude := make(map[string]bool)
	exclude[r.Node.Id] = true

	src := r.OutChannel.Destination
	dst := r.InChannel.Source

	route, err := r.Node.Graph.GetRoute(src, dst, r.Amount, exclude, maxHops)
	if err != nil {
		return nil, err
	}

	route.Prepend(r.OutChannel)
	route.Append(r.InChannel)

	if route.FeePPM() > r.MaxPPM {
		log.Println("best route found was: ", route)
		return nil, util.NewRouteTooExpensiveError(route.FeePPM(), r.MaxPPM)
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
		if err.Error() == util.ErrSendPayTimeout.Error() {
			return nil, util.ErrSendPayTimeout
		}
		return nil, util.ErrTemporaryFailure
	}

	return route, nil
}
