package main

import (
	"github.com/elementsproject/glightning/glightning"
	"github.com/robfig/cron/v3"
	"log"
)

const (
	GRAPH_REFRESH = "5s"
)

type Channel struct {
	glightning.Channel
	Liquidity uint64
}

type Graph struct {
	//map from nodeId to map of channels of that node
	Nodes map[string]map[string]*Channel
}

func (g *Graph) addChannel(c *glightning.Channel) {
	if g.Nodes == nil {
		g.Nodes = make(map[string]map[string]*Channel)
	}
	if g.Nodes[c.Source] == nil {
		g.Nodes[c.Source] = make(map[string]*Channel)
	}
	if g.Nodes[c.Destination] == nil {
		g.Nodes[c.Destination] = make(map[string]*Channel)
	}
	liquidity := estimateInitialLiquidity(c)
	g.Nodes[c.Source][c.Destination] = &Channel{
		Channel:   *c,
		Liquidity: liquidity,
	}
	g.Nodes[c.Destination][c.Source] = &Channel{
		Channel:   *c,
		Liquidity: c.Satoshis - liquidity,
	}
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

func SetRecurrentGraphRefresh(c *cron.Cron, graphRefresh string) {
	_, err := c.AddFunc("@every "+graphRefresh, func() {
		graph = RefreshGraph()
	})
	if err != nil {
		log.Printf("%v\n", err)
	}
}
