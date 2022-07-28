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
	if _, ok := g.Inbound[dst]; !ok {
		log.Println("dst:", util.ErrNoSuchNode)
		return nil, util.ErrNoSuchNode
	}
	if _, ok := g.Inbound[src]; !ok {
		log.Println("src:", util.ErrNoSuchNode)
		return nil, util.ErrNoSuchNode
	}

	// initialize data structures
	distance := make(map[string]int)
	maxDistance := 1 << 31
	for u := range g.Inbound {
		distance[u] = maxDistance
	}
	distance[dst] = 0
	hop := make(map[string]RouteHop)

	// initialize priority queue, put destination in
	pq := make(PriorityQueue, 1, 16)
	pq[0] = &Item{value: &PqItem{
		Node:   dst,
		Amount: amount,
		Delay:  0,
		Hops:   0,
	}, priority: 0}
	heap.Init(&pq)

	// main loop
	for pq.Len() > 0 {
		// get the node with the lowest distance from the priority queue
		pqItem := heap.Pop(&pq).(*Item)
		u := pqItem.value.Node
		amount := pqItem.value.Amount
		delay := pqItem.value.Delay
		hops := pqItem.value.Hops
		priority := pqItem.priority
		// if we already visited this node with a lower distance, ignore it
		if priority > distance[u] {
			continue
		}

		// if we reached the source, we are done
		if u == src {
			break
		}

		// if we reached the maximum number of hops, discard this node
		if hops >= maxHops {
			continue
		}

		// check all the neighbors of the current node
		for v, edge := range g.Inbound[u] {
			if exclude[v] {
				continue
			}

			// for each channel in the edge between two nodes (there may be multiple channels between two nodes)
			for _, scid := range edge {
				channelId := scid + "/" + util.GetDirection(v, u)
				if _, ok := g.Channels[channelId]; !ok {
					log.Println("channel not found:", channelId)
					continue
				}
				channel := g.Channels[channelId]

				// check if the channel is usable
				if !channel.CanForward(amount) {
					continue
				}

				// compute fees and update the priority queue if we found a better way to reach v
				channelFee := channel.ComputeFee(amount)
				newDistance := distance[u] + int(channelFee)
				if newDistance < distance[v] {

					// now v is reachable from u with a lower distance
					distance[v] = newDistance

					// add v to the priority queue while computing fees, delay and hops
					hop[v] = RouteHop{
						channel,
						amount + channelFee,
						delay + channel.Delay,
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
	// if we did not reach the source, we did not find a route
	if distance[src] == maxDistance {
		return nil, util.ErrNoRoute
	}
	
	// now we have the hop map, we can build the hops
	hops := make([]RouteHop, 0, 10)
	for u := src; u != dst; u = hop[u].Destination {
		hops = append(hops, hop[u])
	}
	return hops, nil
}
