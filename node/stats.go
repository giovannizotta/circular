package node

import (
	"circular/graph"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"github.com/elementsproject/glightning/jrpc2"
	"strconv"
	"time"
)

type Stats struct {
	GraphStats *graph.Stats                `json:"graph_stats"`
	Successes  []glightning.SendPaySuccess `json:"successes"`
	Failures   []glightning.SendPayFailure `json:"failures"`
	Routes     []graph.PrettyRoute         `json:"routes"`
}

func (s *Stats) Name() string {
	return "circular-stats"
}

func (s *Stats) New() interface{} {
	return &Stats{}
}

func (s *Stats) Call() (jrpc2.Result, error) {
	return GetNode().GetStats(), nil
}

func (n *Node) GetStats() *Stats {
	defer util.TimeTrack(time.Now(), "node.GetStats", n.Logf)
	n.PeersLock.RLock()
	defer n.PeersLock.RUnlock()

	successes, err := n.DB.ListSuccesses()
	if err != nil {
		n.Logln(glightning.Unusual, err)
	}

	failures, err := n.DB.ListFailures()
	if err != nil {
		n.Logln(glightning.Unusual, err)
	}

	routes, err := n.DB.ListRoutes()
	if err != nil {
		n.Logln(glightning.Unusual, err)
	}

	return &Stats{
		GraphStats: n.Graph.GetStats(),
		Successes:  successes,
		Failures:   failures,
		Routes:     routes,
	}
}

func (s *Stats) String() string {
	var result string
	result += "Node stats:" + "\n"
	result += s.GraphStats.String() + "\n"
	result += "successes: " + strconv.Itoa(len(s.Successes)) + "\n"
	result += "failures: " + strconv.Itoa(len(s.Failures)) + "\n"
	result += "routes: " + strconv.Itoa(len(s.Routes)) + "\n"

	var totalMoved uint64 = 0
	for _, success := range s.Successes {
		totalMoved += success.MilliSatoshi
	}
	result += "Total amount of BTC rebalanced: " + strconv.FormatUint(totalMoved/1000, 10) + "sats"

	return result
}
