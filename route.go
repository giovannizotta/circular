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
	Hops   *[]glightning.RouteHop
}

func NewRoute(in string, out string, amount uint64) (*Route, error) {
	exclude := excludeEdgesToSelf(out)
	exclude = append(exclude, excludeEdgesToSelf(in)...)

	log.Printf("exclude: %v+\n", exclude)
	route, err := lightning.GetRoute(in, amount, 20, 0, out, 5, exclude, 20)
	if err != nil {
		return nil, err
	}

	for i := range route {
		route[i].AmountMsat = ""
	}

	result := &Route{
		In:     in,
		Out:    out,
		Amount: amount,
		Hops:   &route,
	}
	log.Printf("route: %+v\n", result)
	for i := range *result.Hops {
		log.Printf("hop %d: %+v\n", i, (*result.Hops)[i])
	}
	return result, nil
}

func getDirection(from string, to string) uint8 {
	if from < to {
		return 0
	}
	return 1
}

func computeHopFeeMillisatoshi(from string, to string, amount uint64) uint64 {
	baseFee := graph.Nodes[from][to][0].BaseFeeMillisatoshi
	result := baseFee
	proportionalFee := ((amount / 1000) * graph.Nodes[from][to][0].FeePerMillionth) / 1000
	result += proportionalFee
	log.Printf("base fee: %d, proportional fee: %d, result: %d\n", baseFee, proportionalFee, result)
	return result
}

func (r *Route) prependInitialHop(out string) {
	//FIXME: get the best channel?
	routeHop := glightning.RouteHop{
		Id:             out,
		ShortChannelId: graph.Nodes[self.Id][out][0].ShortChannelId,
		MilliSatoshi:   (*r.Hops)[0].MilliSatoshi + computeHopFeeMillisatoshi(self.Id, out, (*r.Hops)[0].MilliSatoshi),
		Delay:          (*r.Hops)[0].Delay + graph.Nodes[self.Id][out][0].Delay,
		Direction:      getDirection(self.Id, out),
	}
	//prepend the hop to the route
	*r.Hops = append([]glightning.RouteHop{routeHop}, *r.Hops...)
}

func (r *Route) appendFinalHop(in string) {
	last := len(*r.Hops) - 1
	routeHop := glightning.RouteHop{
		Id:             self.Id,
		ShortChannelId: graph.Nodes[in][self.Id][0].ShortChannelId,
		MilliSatoshi:   (*r.Hops)[last].MilliSatoshi,
		Delay:          graph.Nodes[in][self.Id][0].Delay,
		Direction:      getDirection(in, self.Id),
	}
	for i := range *r.Hops {
		(*r.Hops)[i].Delay += graph.Nodes[in][self.Id][0].Delay
	}
	*r.Hops = append(*r.Hops, routeHop)

}

func excludeEdgesToSelf(node string) []string {
	var result []string
	for i, channel := range graph.Nodes[self.Id][node] {
		result = append(result, channel.ShortChannelId+"/"+strconv.Itoa(i%2))
	}
	return result
}

func (r *Route) addFirstHopFee() {
	(*r.Hops)[0].MilliSatoshi += computeHopFeeMillisatoshi((*r.Hops)[0].Id, (*r.Hops)[1].Id, (*r.Hops)[0].MilliSatoshi)
}

func (r *Route) sendPay(paymentHash string) (*glightning.SendPayFields, error) {
	_, err := lightning.SendPayLite(*r.Hops, paymentHash)
	if err != nil {
		log.Println(err)
	}

	result, err := lightning.WaitSendPay(paymentHash, 20)
	if err != nil {
		return nil, err
	}
	return result, nil
}
