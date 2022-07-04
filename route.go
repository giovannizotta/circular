package main

import (
	"github.com/elementsproject/glightning/glightning"
	"log"
	"strconv"
)

type Route struct {
	In     string
	Out    string
	Amount uint64
	Hops   []string
}

func NewRoute(in string, out string, amount uint64) (*Route, error) {
	hops, err := buildPath(in, out, amount)
	if err != nil {
		return nil, err
	}

	result := &Route{
		In:     in,
		Out:    out,
		Amount: amount,
		Hops:   hops,
	}
	log.Printf("route: %+v\n", result)
	return result, nil
}

func buildPath(in string, out string, amount uint64) ([]string, error) {
	exclude := excludeEdgesToSelf(out)
	exclude = append(exclude, excludeEdgesToSelf(in)...)

	log.Printf("exclude: %v+\n", exclude)
	route, err := lightning.GetRoute(in, amount, 20, 0, out, 5, exclude, 20)
	if err != nil {
		return nil, err
	}
	var hops []string
	for i := range route {
		hops = append(hops, route[i].Id)
	}
	hops = append([]string{out}, hops...)
	hops = append([]string{self.Id}, hops...)
	hops = append(hops, self.Id)
	return hops, nil
}

func getDirection(from string, to string) uint8 {
	if from < to {
		return 0
	}
	return 1
}

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
	log.Printf("base fee: %d, proportional fee: %d, result: %d\n", baseFee, proportionalFee, result)
	return result
}

func excludeEdgesToSelf(node string) []string {
	var result []string
	for i, channel := range graph.Nodes[self.Id][node] {
		result = append(result, channel.ShortChannelId+"/"+strconv.Itoa(i%2))
	}
	return result
}

func addHop(route *[]glightning.RouteHop, hops []string, i int, amount uint64, delay uint) {
	//TODO: get best channel instead of always using the first one
	if i < 0 {
		return
	}
	from := hops[i]
	to := hops[i+1]
	routeHop := glightning.RouteHop{
		Id:             to,
		ShortChannelId: graph.Nodes[from][to][0].ShortChannelId,
		MilliSatoshi:   amount,
		Delay:          delay,
		Direction:      getDirection(from, to),
	}
	*route = append(*route, routeHop)

	amount += computeFee(from, to, amount)
	delay += graph.Nodes[from][to][0].Delay
	addHop(route, hops, i-1, amount, delay)
}

func reverseRoute(route []glightning.RouteHop) {
	for i := 0; i < len(route)/2; i++ {
		route[i], route[len(route)-i-1] = route[len(route)-i-1], route[i]
	}
}

func (r *Route) toLightningRoute() *[]glightning.RouteHop {
	result := &[]glightning.RouteHop{}
	addHop(result, r.Hops, len(r.Hops)-2, r.Amount, graph.Nodes[r.In][self.Id][0].Delay)
	reverseRoute(*result)
	return result
}
