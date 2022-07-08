package main

import (
	"circular/graph"
	"circular/node"
	"github.com/elementsproject/glightning/glightning"
	"log"
)

func registerOptions(p *glightning.Plugin) {
	err := p.RegisterNewOption("graph_refresh",
		"How often the gossip graph gets refreshed",
		graph.REFRESH_INTERVAL)
	if err != nil {
		log.Fatalln("error registering option graph_refresh:", err)
	}

	err = p.RegisterNewOption("peer_refresh",
		"How often the peer list gets refreshed",
		node.PEER_REFRESH)
	if err != nil {
		log.Fatalln("error registering option peer_refresh:", err)
	}
}
