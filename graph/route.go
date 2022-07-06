package graph

import (
	"github.com/elementsproject/glightning/glightning"
	"log"
)

const (
	DELAY = 6
)

type Route struct {
	Destination string
	Source      string
	Amount      uint64
	Fee         uint64
	FeePPM      uint64
	Hops        []glightning.RouteHop
}

func NewRoute(in string, out string, amount uint64) *Route {
	return &Route{
		Destination: in,
		Source:      out,
		Amount:      amount,
		Fee:         0,
		FeePPM:      0,
	}
}

func getDirection(from string, to string) uint8 {
	if from < to {
		return 0
	}
	return 1
}

//TODO: refactor in "addLastHopDelay" and "addLastHopFee"
// and implement computeDelay and computeFee
func (r *Route) recomputeDelay(lastHopDelay uint) {
	for i := 0; i < len(r.Hops)-1; i++ {
		r.Hops[i].Delay += lastHopDelay
	}
}

func (r *Route) recomputeFee(lastHopFee uint64) {
	log.Printf("lastHopFee: %d", lastHopFee)
	for i := 0; i < len(r.Hops)-1; i++ {
		r.Hops[i].MilliSatoshi += lastHopFee
	}

	r.Fee = r.Hops[0].MilliSatoshi - r.Amount
	r.FeePPM = (r.Fee * 1000000000) / r.Amount
}

func (r *Route) AppendHop(id string, channel *Channel) {
	lastHop := r.Hops[len(r.Hops)-1]

	hop := glightning.RouteHop{
		Id:             id,
		ShortChannelId: channel.ShortChannelId,
		MilliSatoshi:   r.Amount,
		Delay:          DELAY,
		Direction:      getDirection(r.Destination, id),
	}
	r.Hops = append(r.Hops, hop)
	//now add lastHopFee and lastHopDelay
	r.recomputeFee(channel.computeFee(lastHop.MilliSatoshi))
	r.recomputeDelay(DELAY + channel.Delay)
}

func (r *Route) PrependHop(id string, channel *Channel, firstHopChannel *Channel) {
	firstHop := r.Hops[0]
	hop := glightning.RouteHop{
		Id:             r.Source,
		ShortChannelId: channel.ShortChannelId,
		MilliSatoshi:   firstHop.MilliSatoshi + firstHopChannel.computeFee(firstHop.MilliSatoshi),
		Delay:          firstHop.Delay + channel.Delay,
		Direction:      getDirection(id, r.Source),
	}
	r.Hops = append([]glightning.RouteHop{hop}, r.Hops...)
}
