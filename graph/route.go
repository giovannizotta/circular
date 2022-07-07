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

func (r *Route) AppendHop(channel *Channel) {
	n := len(r.Hops) - 1
	hop := r.Hops[n]
	r.Hops = append(r.Hops, channel.GetHop(hop.MilliSatoshi, hop.Delay))

	r.addFee(channel.computeFee(hop.MilliSatoshi), n+1)
	r.addDelay(channel.Delay, n+1)
}

func (r *Route) PrependHop(channel *Channel, firstHopChannel *Channel) {
	n := 0
	hop := r.Hops[n]
	r.Hops = append([]glightning.RouteHop{channel.GetHop(hop.MilliSatoshi, hop.Delay)}, r.Hops...)

	r.addFee(firstHopChannel.computeFee(hop.MilliSatoshi), n+1)
	r.addDelay(channel.Delay, n+1)
}
