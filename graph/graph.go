package graph

import (
	"circular/util"
	"container/heap"
	"errors"
	"github.com/elementsproject/glightning/glightning"
	"log"
	"time"
)

const (
	GRAPH_REFRESH        = "10m"
	FILE                 = "graph.json"
	AVERAGE_AGING_AMOUNT = 0 // the amount by which the liquidity belief is updated
	AGING_VARIANCE       = 0 // the range (+/-) of the random amount added to the liquidity belief
	// for example, for an AVERAGE_AGING_AMOUNT of 10k and an AGING_VARIANCE of 5k
	// the liquidity belief will be updated by a random amount between 5k and 15k	(10k +- 5k)
)

// Edge contains All the SCIDs of the channels going from nodeA to nodeB
type Edge []string

// Graph is the lightning network graph from the perspective of self
// It has been built from the gossip received by lightningd.
// To access the edges flowing out from a node, use: g.Outbound[node]
// To access an edge between nodeA and nodeB, use: g.Outbound[nodeA][nodeB]
// * an edge consists of one or more SCIDs between nodeA and nodeB
// To access a channel via channelId (scid/direction). use: g.Channels[channelId]
type Graph struct {
	Channels map[string]*Channel        `json:"channels"`
	Outbound map[string]map[string]Edge `json:"-"`
	Inbound  map[string]map[string]Edge `json:"-"`
	Aliases  map[string]string          `json:"-"`
}

func NewGraph(filename string) *Graph {
	var g *Graph
	g, err := LoadFromFile(filename)
	if err != nil {
		g = &Graph{
			Channels: make(map[string]*Channel),
			Aliases:  make(map[string]string),
		}
	}
	return g
}

func allocate(links *map[string]map[string]Edge, from, to string) {
	if *links == nil {
		*links = make(map[string]map[string]Edge)
	}
	if (*links)[from] == nil {
		(*links)[from] = make(map[string]Edge)
	}
	if (*links)[from][to] == nil {
		(*links)[from][to] = make([]string, 0)
	}
}

func (g *Graph) AddChannel(c *Channel) {
	allocate(&g.Outbound, c.Source, c.Destination)
	allocate(&g.Inbound, c.Destination, c.Source)
	g.Outbound[c.Source][c.Destination] = append(g.Outbound[c.Source][c.Destination], c.ShortChannelId)
	g.Inbound[c.Destination][c.Source] = append(g.Inbound[c.Destination][c.Source], c.ShortChannelId)
}

func (g *Graph) GetRoute(src, dst string, amount uint64, exclude map[string]bool) (*Route, error) {
	hops, err := g.dijkstra(src, dst, amount, exclude)
	if err != nil {
		return nil, err
	}
	route := NewRoute(src, dst, amount, hops, g)
	return route, nil
}

func (g *Graph) dijkstra(src, dst string, amount uint64, exclude map[string]bool) ([]RouteHop, error) {
	// start from the destination and find the source so that we can compute fees
	// TODO: consider that 32bits fees can be a problem but the api does it in that way
	defer util.TimeTrack(time.Now(), "graph.dijkstra")
	log.Println("looking for a route from", src, "to", dst)
	log.Println("graph has", len(g.Channels), "channels")
	log.Println("graph has", len(g.Inbound), "nodes")
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
	}, priority: 0}
	heap.Init(&pq)

	for pq.Len() > 0 {
		pqItem := heap.Pop(&pq).(*Item)
		u := pqItem.value.Node
		amount := pqItem.value.Amount
		delay := pqItem.value.Delay
		fee := pqItem.priority
		if u == src {
			break
		}
		if fee > distance[u] {
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

				channelFee := int(channel.ComputeFee(amount))
				newDistance := distance[u] + channelFee
				if newDistance < distance[v] {
					distance[v] = newDistance
					hop[v] = RouteHop{
						channel,
						amount,
						delay,
					}
					heap.Push(&pq, &Item{value: &PqItem{
						Node:   v,
						Amount: amount + uint64(channelFee),
						Delay:  delay + channel.Delay,
					}, priority: newDistance})
				}
			}
		}
	}
	if distance[src] == maxDistance {
		log.Println("no route found")
		return nil, errors.New("no route found")
	}
	// now we have the hop map, we can build the hops
	hops := make([]RouteHop, 0, 10)
	for u := src; u != dst; u = hop[u].Channel.Destination {
		hops = append(hops, hop[u])
	}
	return hops, nil
}

func (g *Graph) RefreshChannels(channelList []*glightning.Channel) {
	// we need to do NewChannel and not only update the liquidity because of gossip updates
	defer util.TimeTrack(time.Now(), "graph.RefreshChannels")
	for _, c := range channelList {
		var channel *Channel
		channelId := c.ShortChannelId + "/" + util.GetDirection(c.Source, c.Destination)
		// if the channel did not exist prior to this refresh estimate its initial liquidity to be 50/50
		perfectBalance := uint64(0.5 * float64(c.Satoshis*1000))
		if _, ok := g.Channels[channelId]; !ok {
			channel = NewChannel(c, perfectBalance)
			g.AddChannel(channel)
		} else {
			liquidity := g.getLiquidityAfterAging(channelId, perfectBalance)
			channel = NewChannel(c, liquidity)
			g.updateOppositeChannel(channel, liquidity)
		}
		g.Channels[channelId] = channel
	}
}

func (g *Graph) getLiquidityAfterAging(channelId string, perfectBalance uint64) uint64 {
	aging := util.RandRange(AVERAGE_AGING_AMOUNT-AGING_VARIANCE, AVERAGE_AGING_AMOUNT+AGING_VARIANCE)
	return util.Max(g.Channels[channelId].Liquidity+aging, perfectBalance)
}

func (g *Graph) updateOppositeChannel(c *Channel, liquidity uint64) {
	oppositeChannelId := c.ShortChannelId + "/" + util.GetDirection(c.Destination, c.Source)
	// if opposite channel is in the map
	if _, ok := g.Channels[oppositeChannelId]; ok {
		oppositeChannel := g.Channels[oppositeChannelId]
		oppositeChannel.Liquidity = (c.Satoshis * 1000) - liquidity
	}
}

func (g *Graph) RefreshAliases(nodes []*glightning.Node) {
	for _, n := range nodes {
		g.Aliases[n.Id] = n.Alias
	}
}
