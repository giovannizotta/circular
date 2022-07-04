package main

import (
	"github.com/elementsproject/glightning/glightning"
)

type Route struct {
	In     string
	Out    string
	Amount uint64
	Hops   []string
}

func NewRoute(in string, out string, amount uint64, exclude []string) (*Route, error) {
	hops, err := buildPath(in, out, amount, exclude)
	if err != nil {
		return nil, err
	}

	result := &Route{
		In:     in,
		Out:    out,
		Amount: amount,
		Hops:   hops,
	}
	return result, nil
}

func buildPath(in string, out string, amount uint64, exclude []string) ([]string, error) {
	route, err := lightning.GetRoute(in, amount, 20, 0, out, 5, exclude, 20)
	if err != nil {
		return nil, err
	}
	var hops []string
	for i := range route {
		hops = append(hops, route[i].Id)
	}

	hops = append([]string{out}, hops...)
	return hops, nil
}

func getDirection(from string, to string) uint8 {
	if from < to {
		return 0
	}
	return 1
}

//actually returns satoshi per Billion, not per Million
func getRoutePPM(route []glightning.RouteHop) uint64 {
	originalAmount := route[len(route)-1].MilliSatoshi
	fee := (route[0].MilliSatoshi) - originalAmount
	return (fee * 1000000000) / originalAmount
}

func computeFee(from string, to string, amount uint64) uint64 {
	baseFee := graph.Nodes[from][to][0].BaseFeeMillisatoshi
	result := baseFee
	proportionalFee := ((amount / 1000) * graph.Nodes[from][to][0].FeePerMillionth) / 1000
	result += proportionalFee
	return result
}

func reverseRoute(route []glightning.RouteHop) {
	for i := 0; i < len(route)/2; i++ {
		route[i], route[len(route)-i-1] = route[len(route)-i-1], route[i]
	}
}

func (r *Route) toLightningRoute() *[]glightning.RouteHop {
	//TODO: get best channel instead of always using the first one
	result := &[]glightning.RouteHop{}
	amount := r.Amount
	lastId := r.Hops[len(r.Hops)-1]
	delay := graph.Nodes[r.In][lastId][0].Delay
	for i := len(r.Hops) - 2; i >= 0; i-- {
		from := r.Hops[i]
		to := r.Hops[i+1]
		routeHop := glightning.RouteHop{
			Id:             to,
			ShortChannelId: graph.Nodes[from][to][0].ShortChannelId,
			MilliSatoshi:   amount,
			Delay:          delay,
			Direction:      getDirection(from, to),
		}
		amount += computeFee(from, to, amount)
		delay += graph.Nodes[from][to][0].Delay
		*result = append(*result, routeHop)
	}
	reverseRoute(*result)
	return result
}
