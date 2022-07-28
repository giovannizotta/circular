package graph

import (
	"fmt"
	"strconv"
)

type PrettyRouteHop struct {
	Id             string `json:"id"`
	Alias          string `json:"alias"`
	ShortChannelId string `json:"short_channel_id"`
	MilliSatoshi   uint64 `json:"millisatoshi"`
	Delay          uint   `json:"delay"`
	Fee            uint64 `json:"fee"`
	FeePPM         uint64 `json:"ppm"`
}

type PrettyRoute struct {
	SourceId         string           `json:"source_id"`
	DestinationId    string           `json:"destination_id"`
	SourceAlias      string           `json:"source_alias"`
	DestinationAlias string           `json:"destination_alias"`
	Amount           uint64           `json:"amount_sat"`
	Fee              uint64           `json:"fee_msat"`
	FeePPM           uint64           `json:"ppm"`
	Hops             []PrettyRouteHop `json:"hops"`
}

func NewPrettyRoute(route *Route) *PrettyRoute {
	hops := make([]PrettyRouteHop, len(route.Hops))

	// now hops
	from := route.Hops[0].Source
	hops[0] = PrettyRouteHop{
		Id:             from,
		ShortChannelId: route.Hops[0].ShortChannelId,
		MilliSatoshi:   route.Hops[0].MilliSatoshi,
		Delay:          route.Hops[0].Delay,
		Fee:            0,
		FeePPM:         0,
	}

	hops[0].Alias = route.Graph.GetAlias(from)

	for i := 1; i < len(route.Hops); i++ {
		fee := route.Hops[i-1].MilliSatoshi - route.Hops[i].MilliSatoshi
		feePPM := fee * 1000000 / route.Hops[i].MilliSatoshi
		from = route.Hops[i].Source
		hops[i] = PrettyRouteHop{
			Id:             from,
			ShortChannelId: route.Hops[i].ShortChannelId,
			MilliSatoshi:   route.Hops[i].MilliSatoshi,
			Delay:          route.Hops[i].Delay,
			Fee:            fee,
			FeePPM:         feePPM,
		}
		hops[i].Alias = route.Graph.GetAlias(from)
	}

	return &PrettyRoute{
		SourceId:         route.Source,
		DestinationId:    route.Destination,
		SourceAlias:      route.Graph.GetAlias(route.Source),
		DestinationAlias: route.Graph.GetAlias(route.Destination),
		Amount:           route.Amount / 1000,
		Fee:              route.Fee(),
		FeePPM:           route.FeePPM(),
		Hops:             hops,
	}
}

func (r *PrettyRoute) String() string {
	var result string
	result += "Route from: " + r.SourceAlias + " to: " + r.DestinationAlias + "\n"
	result += "Amount: " + strconv.FormatUint(r.Amount, 10) + "\n"
	result += "Fee: " + strconv.FormatUint(r.Fee, 10) + "msat\n"
	result += "Fee PPM: " + strconv.FormatUint(r.FeePPM, 10) + "\n"
	result += "Hops: " + strconv.Itoa(len(r.Hops)) + "\n"

	for i := 0; i < len(r.Hops); i++ {
		alias := r.Hops[i].Alias
		fee := r.Hops[i].Fee
		feePPM := r.Hops[i].FeePPM
		delay := r.Hops[i].Delay
		shortChannelId := r.Hops[i].ShortChannelId

		result += fmt.Sprintf("Hop %2d: %40s, fee: %8.3f, ppm: %5d, scid: %s, delay: %d\n",
			i+1, alias,
			float64(fee)/1000, feePPM,
			shortChannelId, delay)
	}
	return result
}

func (r *PrettyRoute) Simple() string {
	var result string
	result += "Sending " + strconv.FormatUint(r.Amount, 10) + " sats from [" + r.SourceAlias + "] to [" + r.DestinationAlias
	result += "] over " + strconv.Itoa(len(r.Hops)) + " hops, costing " + strconv.FormatUint(r.Fee, 10) + "msat ("
	result += strconv.FormatUint(r.FeePPM, 10) + "PPM)"
	result += " via "
	for i := 0; i < len(r.Hops); i++ {
		alias := r.Hops[i].Alias
		feePPM := r.Hops[i].FeePPM
		result += "- " + alias + " (" + strconv.FormatUint(feePPM, 10) + "PPM) "
	}
	return result
}
