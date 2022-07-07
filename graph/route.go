package graph

import (
	"github.com/elementsproject/glightning/glightning"
)

type Route struct {
	Destination string
	Source      string
	Amount      uint64
	Hops        []glightning.RouteHop
}

func NewRoute(in string, out string, amount uint64, hops []glightning.RouteHop) *Route {
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

func (r *Route) addDelay(lastHopDelay uint) {
	for i := 0; i < len(r.Hops)-1; i++ {
		r.Hops[i].Delay += lastHopDelay
	}
}

func (r *Route) addFee(lastHopFee uint64) {
	for i := 0; i < len(r.Hops)-1; i++ {
		r.Hops[i].MilliSatoshi += lastHopFee
	}
}

func (r *Route) AppendHop(channel *Channel) {
	lastHop := r.Hops[len(r.Hops)-1]
	r.Hops = append(r.Hops, channel.GetHop(r.Amount, 0))

	r.addFee(channel.computeFee(lastHop.MilliSatoshi))
	r.addDelay(channel.Delay)
}

func (r *Route) PrependHop(channel *Channel, firstHopChannel *Channel) {
	firstHop := r.Hops[0]
	r.Hops = append([]glightning.RouteHop{channel.GetHop(firstHop.MilliSatoshi, firstHop.Delay)}, r.Hops...)

	r.addFee(firstHopChannel.computeFee(firstHop.MilliSatoshi))
	r.addDelay(channel.Delay)
}
