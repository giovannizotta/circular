package main

import (
	"container/heap"
	"errors"
	"github.com/elementsproject/glightning/glightning"
	"log"
)

type Route struct {
	Destination string
	Source      string
	Amount      uint64
	Fee         uint64
	FeePPM      uint64
	Hops        []glightning.RouteHop
}

func NewRoute(in string, out string, amount uint64, exclude []string) (*Route, error) {
	excludeMap := make(map[string]bool)
	for _, v := range exclude {
		excludeMap[v] = true
	}

	hops, err := getRoute(out, in, amount, excludeMap)
	if err != nil {
		return nil, err
	}

	return &Route{
		Destination: in,
		Source:      out,
		Amount:      amount,
		Fee:         0,
		FeePPM:      0,
		Hops:        hops,
	}, nil
}

func getRoute(src, dst string, amount uint64, exclude map[string]bool) ([]glightning.RouteHop, error) {
	// start from the destination and find the source so that we can compute fees
	src, dst = dst, src
	distance := make(map[string]int)
	parent := make(map[string]glightning.RouteHop)
	maxDistance := 1 << 31
	for u := range graph.Outbound {
		distance[u] = maxDistance
	}
	pq := make(PriorityQueue, 1, 16)
	// Insert source and give it a priority of 0
	pq[0] = &Item{value: &PqItem{
		Node:   src,
		Amount: amount,
		Delay:  0,
	}, priority: 0}
	distance[src] = 0
	heap.Init(&pq)

	for pq.Len() > 0 {
		pqItem := heap.Pop(&pq).(*Item)
		u := pqItem.value.Node
		amount := pqItem.value.Amount
		delay := pqItem.value.Delay
		fee := pqItem.priority
		if u == dst {
			break
		}
		if fee > distance[u] {
			continue
		}
		for v, edge := range graph.Inbound[u] {
			if exclude[v] {
				continue
			}
			for scid, channel := range edge {
				channelFee := int(channel.computeFee(amount))
				newDistance := distance[u] + channelFee
				if newDistance < distance[v] {
					distance[v] = newDistance
					parent[v] = glightning.RouteHop{
						Id:             u,
						ShortChannelId: scid,
						MilliSatoshi:   amount,
						Delay:          delay,
						Direction:      getDirection(v, u),
					}
					heap.Push(&pq, &Item{value: &PqItem{
						Node:   v,
						Amount: amount + uint64(channelFee),
						Delay:  delay + channel.Delay,
					}, priority: channelFee})
				}
			}
		}
	}
	if distance[dst] == maxDistance {
		return nil, errors.New("no route found")
	}
	// now we have the parent map, we can build the hops
	hops := make([]glightning.RouteHop, 0, 10)
	for u := dst; u != src; u = parent[u].Id {
		hops = append(hops, parent[u])
	}
	log.Println("hops before adding self")
	for i, hop := range hops {
		log.Printf("hop %d: %+v", i, hop)
	}

	return hops, nil
}

func getDirection(from string, to string) uint8 {
	if from < to {
		return 0
	}
	return 1
}

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

func (r *Route) appendHop(id string) {
	lastHop := r.Hops[len(r.Hops)-1]
	bestScid := getBestPeerChannel(r.Destination, func(channel *glightning.PeerChannel) uint64 {
		return channel.ReceivableMilliSatoshi
	}).ShortChannelId
	channel := graph.Outbound[r.Destination][id][bestScid]

	hop := glightning.RouteHop{
		Id:             id,
		ShortChannelId: channel.ShortChannelId,
		MilliSatoshi:   r.Amount,
		Delay:          0,
		Direction:      getDirection(r.Destination, id),
	}
	r.Hops = append(r.Hops, hop)
	//now add lastHopFee and lastHopDelay
	r.recomputeFee(channel.computeFee(lastHop.MilliSatoshi))
	r.recomputeDelay(channel.Delay)
}

func (r *Route) prependHop(id string) {
	firstHop := r.Hops[0]
	bestScid := getBestPeerChannel(r.Source, func(channel *glightning.PeerChannel) uint64 {
		return channel.SpendableMilliSatoshi
	}).ShortChannelId
	channel := graph.Outbound[id][r.Source][bestScid]
	outToFirstHop := graph.Outbound[r.Source][firstHop.Id][firstHop.ShortChannelId]

	hop := glightning.RouteHop{
		Id:             r.Source,
		ShortChannelId: channel.ShortChannelId,
		MilliSatoshi:   firstHop.MilliSatoshi + outToFirstHop.computeFee(firstHop.MilliSatoshi),
		Delay:          firstHop.Delay + channel.Delay,
		Direction:      getDirection(id, r.Source),
	}
	r.Hops = append([]glightning.RouteHop{hop}, r.Hops...)
}
