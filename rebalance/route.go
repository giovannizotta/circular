package rebalance

import (
	"circular/graph"
	"circular/node"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"time"
)

func (r *Rebalance) getRoute(maxHops int) (*graph.Route, error) {
	defer util.TimeTrack(time.Now(), "rebalance.getRoute", r.Node.Logf)
	exclude := make(map[string]bool)
	exclude[r.Node.Id] = true

	src := r.OutChannel.Destination
	dst := r.InChannel.Source

	r.Node.Logln(glightning.Debug, "looking for a route from ", r.Node.Graph.GetAlias(src), " to ", r.Node.Graph.GetAlias(dst))
	route, err := r.Node.Graph.GetRoute(src, dst, r.Amount, exclude, maxHops)
	if err != nil {
		return nil, err
	}

	route.Prepend(r.OutChannel)
	route.Append(r.InChannel)

	if route.FeePPM() > r.MaxPPM {
		return nil, util.NewRouteTooExpensiveError(route.FeePPM(), r.MaxPPM)
	}

	return route, nil
}

func (r *Rebalance) tryRoute(maxHops int) (*graph.PrettyRoute, error) {
	paymentSecretHash, err := r.Node.GeneratePreimageHashPair()
	if err != nil {
		return nil, err
	}

	r.Node.Logln(glightning.Debug, "generating route")
	route, err := r.getRoute(maxHops)
	if err != nil {
		return nil, err
	}

	prettyRoute := graph.NewPrettyRoute(route, paymentSecretHash)

	// save route to DB
	if err := r.Node.SaveToDb(node.ROUTE_PREFIX+paymentSecretHash, prettyRoute); err != nil {
		r.Node.Logln(glightning.Unusual, "unable to save route to db: ", err)
	}
	r.Node.Logln(glightning.Debug, prettyRoute)
	r.Node.Logln(glightning.Info, prettyRoute.Simple())

	_, err = r.Node.SendPay(route, paymentSecretHash)
	if err != nil {
		if err == util.ErrSendPayTimeout {
			return nil, err
		}
		if err == util.ErrWireFeeInsufficient {
			return nil, err
		}
		if err == util.ErrFirstPeerNotReady {
			return nil, err
		}
		return nil, util.ErrTemporaryFailure
	}

	return prettyRoute, nil
}
