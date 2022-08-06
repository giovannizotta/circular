package node

import (
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"github.com/elementsproject/glightning/jrpc2"
	"time"
)

type DeleteStats struct {
	Failures  int `json:"failures"`
	Successes int `json:"successes"`
	Routes    int `json:"routes"`
}

func (s *DeleteStats) Name() string {
	return "circular-delete-stats"
}

func (s *DeleteStats) New() interface{} {
	return &DeleteStats{}
}

func (s *DeleteStats) Call() (jrpc2.Result, error) {
	return GetNode().DeleteStats(), nil
}

func (n *Node) DeleteStats() *DeleteStats {
	defer util.TimeTrack(time.Now(), "node.DeleteStats", n.Logf)

	successes, err := n.DB.DeleteSuccesses()
	if err != nil {
		n.Logln(glightning.Unusual, err)
	}
	failures, err := n.DB.DeleteFailures()
	if err != nil {
		n.Logln(glightning.Unusual, err)
	}
	routes, err := n.DB.DeleteRoutes()
	if err != nil {
		n.Logln(glightning.Unusual, err)
	}

	return &DeleteStats{
		Successes: successes,
		Failures:  failures,
		Routes:    routes,
	}
}
