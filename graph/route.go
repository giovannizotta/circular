package graph

import (
	"fmt"
	"github.com/elementsproject/glightning/glightning"
	"strconv"
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
	Graph       *Graph
}

func NewRoute(in string, out string, amount uint64, hops []RouteHop, graph *Graph) *Route {
	return &Route{
		Destination: in,
		Source:      out,
		Amount:      amount,
		Hops:        hops,
		Graph:       graph,
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
	newFirstHop := RouteHop{
		Channel:      channel,
		MilliSatoshi: firstHop.MilliSatoshi,
		Delay:        firstHop.Delay,
	}
	r.Hops = append([]RouteHop{newFirstHop}, r.Hops...)
}

func (r *Route) recomputeFeeAndDelay() {
	for i := len(r.Hops) - 2; i >= 0; i-- {
		hop := r.Hops[i+1]
		amountToForward := hop.MilliSatoshi
		r.Hops[i].MilliSatoshi = amountToForward + hop.ComputeFee(amountToForward)

		delay := hop.Delay
		r.Hops[i].Delay = delay + hop.Channel.Delay
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
			Id:             hop.Destination,
			ShortChannelId: hop.ShortChannelId,
			MilliSatoshi:   hop.MilliSatoshi,
			Delay:          hop.Delay,
			Direction:      hop.GetDirection(),
		})
	}
	return hops
}

func (r *Route) String() string {
	var result string
	result += "Route from: " + r.Source[:8] + " to: " + r.Destination[:8] + "\n"
	result += "Amount: " + strconv.FormatUint(r.Amount/1000, 10) + "\n"
	result += "Fee: " + strconv.FormatUint(r.Fee(), 10) + "msat\n"
	result += "Fee PPM: " + strconv.FormatUint(r.FeePPM(), 10) + "\n"
	result += "Hops: " + strconv.Itoa(len(r.Hops)) + "\n"

	from := r.Hops[0].Source
	if alias, ok := r.Graph.Aliases[from]; ok {
		from = alias
	}

	result += fmt.Sprintf("Hop %2d: %40s, fee: %8.3f, ppm: %5d, scid: %s, delay: %d\n",
		1, from, 0.0, 0, r.Hops[0].ShortChannelId, r.Hops[0].Delay)
	for i := 1; i < len(r.Hops); i++ {
		fee := r.Hops[i-1].MilliSatoshi - r.Hops[i].MilliSatoshi
		feePPM := fee * 1000000 / r.Hops[i].MilliSatoshi
		from = r.Hops[i].Source
		if alias, ok := r.Graph.Aliases[from]; ok {
			from = alias
		}
		result += fmt.Sprintf("Hop %2d: %40s, fee: %8.3f, ppm: %5d, scid: %s, delay: %d\n",
			i+1, from,
			float64(fee)/1000, feePPM,
			r.Hops[i].ShortChannelId, r.Hops[i].Delay)
	}
	return result
}
