package graph

import (
	"github.com/elementsproject/glightning/glightning"
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
	return (r.Fee() * 1000000000) / r.Amount
}

func (r *Route) addDelay(delay uint, upTo int) {
	for i := 0; i < upTo; i++ {
		r.Hops[i].Delay += delay
	}
}

func (r *Route) addFee(fee uint64, upTo int) {
	for i := 0; i < upTo; i++ {
		r.Hops[i].MilliSatoshi += fee
	}
}

func getNewHop(channel *Channel, lastHop RouteHop) RouteHop {
	newHop := RouteHop{
		Channel:      channel,
		MilliSatoshi: lastHop.MilliSatoshi,
		Delay:        lastHop.Delay,
	}
	return newHop
}

func (r *Route) add(channel *Channel, where int, f func(hop RouteHop) []RouteHop) {
	targetHop := r.Hops[where]
	r.Hops = f(getNewHop(channel, targetHop))
	r.addFee(channel.computeFee(targetHop.MilliSatoshi), where+1)
	r.addDelay(channel.Delay, where+1)
}

func (r *Route) Prepend(channel *Channel) {
	r.add(channel, 0, func(hop RouteHop) []RouteHop {
		return append([]RouteHop{hop}, r.Hops...)
	})
}

func (r *Route) Append(channel *Channel) {
	r.add(channel, len(r.Hops)-1, func(hop RouteHop) []RouteHop {
		return append(r.Hops, hop)
	})
}

func (r *Route) ToLightningRoute() []glightning.RouteHop {
	var hops []glightning.RouteHop
	for _, hop := range r.Hops {
		hops = append(hops, glightning.RouteHop{
			Id:             hop.Channel.Destination,
			ShortChannelId: hop.Channel.ShortChannelId,
			MilliSatoshi:   hop.MilliSatoshi,
			Delay:          hop.Delay,
			Direction:      hop.Channel.getDirection(),
		})
	}
	return hops
}
