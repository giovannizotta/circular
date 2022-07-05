package main

import (
	"github.com/elementsproject/glightning/glightning"
	"log"
	"strconv"
	"strings"
)

const (
	GRAPH_REFRESH = "10m"
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

func all(v []bool) bool {
	for _, b := range v {
		if !b {
			return false
		}
	}
	return true
}

func (c *Channel) canUse(amount uint64) bool {
	maxHtlcMillisat, _ := strconv.ParseUint(strings.TrimSuffix(c.HtlcMaximumMilliSatoshis, "msat"), 10, 64)
	conditions := []bool{
		c.Liquidity >= amount,
		c.IsActive,
		maxHtlcMillisat >= amount,
	}
	return all(conditions)
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
	return uint64(0.5 * float64(c.Satoshis*1000))
}

func RefreshGraph() *Graph {
	//TODO: persistency
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
