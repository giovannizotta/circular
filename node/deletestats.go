package node

import (
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"github.com/elementsproject/glightning/jrpc2"
	"time"
)

type DeleteStats struct {
	Status string `json:"status"`
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

	if err := n.DB.db.DropPrefix(
		[]byte(SUCCESS_PREFIX),
		[]byte(FAILURE_PREFIX),
		[]byte(ROUTE_PREFIX)); err != nil {
		n.Logf(glightning.Unusual, "Error dropping stats: %v", err)
		return &DeleteStats{Status: "failure"}
	}

	return &DeleteStats{
		Status: "success",
	}
}
