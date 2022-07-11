package graph

import (
	"github.com/elementsproject/glightning/glightning"
)

const (
	INITIAL_DELAY = 144
)

type RouteHop struct {
	*Channel
	MilliSatoshi uint64
	Delay        uint
}

type Route struct {
	Destination string
	Source      string
	Amount      uint64
	Hops        []RouteHop
}

func NewRoute(in string, out string, amount uint64, hops []RouteHop) *Route {
	return &Route{
		Destination: in,
		Source:      out,
		Amount:      amount,
		Hops:        hops,
	}
}

func (r *Route) Fee() uint64 {
	return r.Hops[0].MilliSatoshi - r.Amount
}

func (r *Route) FeePPM() uint64 {
	return (r.Fee() * 1000000) / r.Amount
}

func (r *Route) Prepend(channel *Channel) {
	firstHop := r.Hops[0]
	firstHop.MilliSatoshi += firstHop.Channel.ComputeFee(firstHop.MilliSatoshi)
	newFirstHop := RouteHop{
		Channel:      channel,
		MilliSatoshi: firstHop.MilliSatoshi,
		Delay:        firstHop.Delay + firstHop.Channel.Delay,
	}
	r.Hops = append([]RouteHop{newFirstHop}, r.Hops...)
}

func (r *Route) recomputeFeeAndDelay() {
	for i := len(r.Hops) - 2; i >= 0; i-- {
		nextHop := r.Hops[i+1]
		amountToForward := nextHop.MilliSatoshi
		delay := nextHop.Delay
		r.Hops[i].MilliSatoshi = amountToForward + nextHop.Channel.ComputeFee(amountToForward)
		r.Hops[i].Delay = delay + nextHop.Channel.Delay
	}
}

func (r *Route) Append(channel *Channel) {
	newLastHop := RouteHop{
		Channel:      channel,
		MilliSatoshi: r.Amount,
		Delay:        INITIAL_DELAY,
	}
	r.Hops = append(r.Hops, newLastHop)
	r.recomputeFeeAndDelay()
}

func (r *Route) ToLightningRoute() []glightning.RouteHop {
	var hops []glightning.RouteHop
	for _, hop := range r.Hops {
		hops = append(hops, glightning.RouteHop{
			Id:             hop.Channel.Destination,
			ShortChannelId: hop.Channel.ShortChannelId,
			MilliSatoshi:   hop.MilliSatoshi,
			Delay:          hop.Delay,
			Direction:      hop.Channel.GetDirection(),
		})
	}
	return hops
}
