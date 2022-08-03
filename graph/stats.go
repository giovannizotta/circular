package graph

import (
	"strconv"
	"strings"
)

type Stats struct {
	Nodes           int `json:"nodes"`
	Channels        int `json:"channels"`
	ActiveChannels  int `json:"active_channels"`
	LiquidChannels  int `json:"liquid_channels"`
	MaxHtlcChannels int `json:"max_htlc_channels"`
}

func (g *Graph) GetStats() *Stats {
	g.channelsLock.RLock()
	g.adjacencyListLock.RLock()
	defer g.adjacencyListLock.RUnlock()
	defer g.channelsLock.RUnlock()

	activeChannels := 0
	atLeast200kLiquidity := 0
	atLeast200kMaxHtlc := 0
	for _, c := range g.Channels {
		if c.IsActive {
			activeChannels++
		}
		if c.Liquidity >= 200000000 {
			atLeast200kLiquidity++
		}
		maxHtlc, _ := strconv.ParseUint(strings.TrimSuffix(c.HtlcMaximumMilliSatoshis, "msat"), 10, 64)
		if maxHtlc >= 200000000 {
			atLeast200kMaxHtlc++
		}
	}

	return &Stats{
		Nodes:           len(g.Inbound),
		Channels:        len(g.Channels),
		ActiveChannels:  activeChannels,
		LiquidChannels:  atLeast200kLiquidity,
		MaxHtlcChannels: atLeast200kMaxHtlc,
	}
}

func (s *Stats) String() string {
	var result string
	result += "Graph stats:\n"
	result += "graph has " + strconv.Itoa(s.Nodes) + " nodes\n"
	result += "graph has " + strconv.Itoa(s.Channels) + " channels\n"
	result += "graph has " + strconv.Itoa(s.ActiveChannels) + " active channels\n"
	result += "graph has " + strconv.Itoa(s.LiquidChannels) + " channels believed to have at least 200k liquidity\n"
	result += "graph has " + strconv.Itoa(s.MaxHtlcChannels) + " channels with at least 200k max htlc"
	return result
}
