package graph

import (
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
