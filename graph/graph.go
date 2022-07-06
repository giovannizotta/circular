package graph

import (
	"container/heap"
	"errors"
	"github.com/elementsproject/glightning/glightning"
)

const (
	GRAPH_REFRESH = "10m"
)

// ShortChannelId -> Channel
type Edge map[string]Channel

// id -> id -> Edges
type Graph struct {
	Outbound map[string]map[string]Edge
	Inbound  map[string]map[string]Edge
}

// constructor
func NewGraph() *Graph {
	return &Graph{}
}

func allocate(links *map[string]map[string]Edge, from, to string) {
	if *links == nil {
		*links = make(map[string]map[string]Edge)
	}
	if (*links)[from] == nil {
		(*links)[from] = make(map[string]Edge)
	}
	if (*links)[from][to] == nil {
		(*links)[from][to] = make(Edge)
	}
}

func (g *Graph) AddChannel(c *glightning.Channel) {
	allocate(&g.Outbound, c.Source, c.Destination)
	allocate(&g.Inbound, c.Destination, c.Source)
	liquidity := estimateInitialLiquidity(c)
	(g.Outbound)[c.Source][c.Destination][c.ShortChannelId] =
		Channel{*c, liquidity}
	(g.Inbound)[c.Destination][c.Source][c.ShortChannelId] =
		Channel{*c, c.Satoshis - liquidity}
}

func estimateInitialLiquidity(c *glightning.Channel) uint64 {
	return uint64(0.5 * float64(c.Satoshis*1000))
}

func (g *Graph) GetRoute(src, dst string, amount uint64, exclude map[string]bool) (*Route, error) {
	hops, err := g.dijkstra(src, dst, amount, exclude)
	if err != nil {
		return nil, err
	}
	route := NewRoute(dst, src, amount)
	route.Hops = hops
	return route, nil
}

func (g *Graph) dijkstra(src, dst string, amount uint64, exclude map[string]bool) ([]glightning.RouteHop, error) {
	// start from the destination and find the source so that we can compute fees
	// TODO: consider that 32bits fees can be a problem
	// but the api does it in that way
	src, dst = dst, src
	distance := make(map[string]int)
	parent := make(map[string]glightning.RouteHop)
	maxDistance := 1 << 31
	for u := range g.Outbound {
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
		for v, edge := range g.Inbound[u] {
			if exclude[v] {
				continue
			}
			for scid, channel := range edge {
				if !channel.canUse(amount) {
					continue
				}
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

	return hops, nil
}
