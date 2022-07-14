package graph

import (
	"circular/util"
	"container/heap"
	"log"
	"time"
)

func (g *Graph) GetRoute(src, dst string, amount uint64, exclude map[string]bool, maxHops int) (*Route, error) {
	hops, err := g.dijkstra(src, dst, amount, exclude, maxHops-2) // -2 because we already know the source and destination
	if err != nil {
		return nil, err
	}
	route := NewRoute(src, dst, amount, hops, g)
	return route, nil
}

func (g *Graph) dijkstra(src, dst string, amount uint64, exclude map[string]bool, maxHops int) ([]RouteHop, error) {
	// start from the destination and find the source so that we can compute fees
	// TODO: consider that 32bits fees can be a problem but the api does it in that way
	defer util.TimeTrack(time.Now(), "graph.dijkstra")
	log.Println("looking for a route from", src, "to", dst)
	distance := make(map[string]int)
	hop := make(map[string]RouteHop)
	maxDistance := 1 << 31
	for u := range g.Inbound {
		distance[u] = maxDistance
	}
	distance[dst] = 0

	pq := make(PriorityQueue, 1, 16)
	// Insert destination
	pq[0] = &Item{value: &PqItem{
		Node:   dst,
		Amount: amount,
		Delay:  0,
		Hops:   0,
	}, priority: 0}
	heap.Init(&pq)

	for pq.Len() > 0 {
		pqItem := heap.Pop(&pq).(*Item)
		u := pqItem.value.Node
		amount := pqItem.value.Amount
		delay := pqItem.value.Delay
		hops := pqItem.value.Hops
		priority := pqItem.priority
		if u == src {
			break
		}
		if priority > distance[u] {
			continue
		}
		if hops >= maxHops {
			continue
		}
		for v, edge := range g.Inbound[u] {
			if exclude[v] {
				continue
			}
			for _, scid := range edge {
				channel := g.Channels[scid+"/"+util.GetDirection(v, u)]
				if !channel.CanUse(amount) {
					continue
				}
				channelFee := channel.ComputeFee(amount)
				newDistance := distance[u] + int(channelFee)
				if newDistance < distance[v] {
					distance[v] = newDistance
					hop[v] = RouteHop{
						channel,
						amount,
						delay,
					}
					heap.Push(&pq, &Item{value: &PqItem{
						Node:   v,
						Amount: amount + channelFee,
						Delay:  delay + channel.Delay,
						Hops:   hops + 1,
					}, priority: newDistance})
				}
			}
		}
	}
	if distance[src] == maxDistance {
		log.Println(ErrNoRoute)
		return nil, ErrNoRoute
	}
	// now we have the hop map, we can build the hops
	hops := make([]RouteHop, 0, 10)
	for u := src; u != dst; u = hop[u].Channel.Destination {
		hops = append(hops, hop[u])
	}
	return hops, nil
}
