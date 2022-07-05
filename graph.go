package main

import (
	"github.com/elementsproject/glightning/glightning"
	"log"
)

const (
	GRAPH_REFRESH = "5s"
)

type Channel struct {
	glightning.Channel
	Liquidity uint64
}

func (c *Channel) computeFee(amount uint64) uint64 {
	baseFee := c.BaseFeeMillisatoshi
	result := baseFee
	proportionalFee := ((amount / 1000) * c.FeePerMillionth) / 1000
	result += proportionalFee
	return result
}

// ShortChannelId -> Channel
type Edge map[string]Channel

// id -> id -> Edges
type Graph struct {
	Outbound map[string]map[string]Edge
	Inbound  map[string]map[string]Edge
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

func (g *Graph) addChannel(c *glightning.Channel) {
	allocate(&g.Outbound, c.Source, c.Destination)
	allocate(&g.Inbound, c.Destination, c.Source)
	liquidity := estimateInitialLiquidity(c)
	g.Outbound[c.Source][c.Destination][c.ShortChannelId] =
		Channel{*c, liquidity}
	g.Inbound[c.Destination][c.Source][c.ShortChannelId] =
		Channel{*c, c.Satoshis - liquidity}
}

func estimateInitialLiquidity(c *glightning.Channel) uint64 {
	return uint64(0.5 * float64(c.Satoshis))
}

func RefreshGraph() *Graph {
	log.Println("refreshing graph")
	newGraph := &Graph{}

	channelList, err := lightning.ListChannels()
	if err != nil {
		log.Printf("error listing channels: %v\n", err)
	}

	for _, c := range channelList {
		newGraph.addChannel(c)
	}
	return newGraph
}
