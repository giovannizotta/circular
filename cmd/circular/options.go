package main

import (
	"circular/graph"
	"circular/node"
	"github.com/elementsproject/glightning/glightning"
	"log"
)

func registerOptions(p *glightning.Plugin) {
	if err := p.RegisterNewIntOption("circular-graph-refresh",
		"How often the gossip graph gets refreshed (minutes)",
		graph.DEFAULT_GRAPH_REFRESH_INTERVAL); err != nil {

		log.Fatalln("error registering option circular-graph-refresh:", err)
	}

	if err := p.RegisterNewIntOption("circular-peer-refresh",
		"How often the peer list gets refreshed (seconds)",
		node.DEFAULT_PEER_REFRESH_INTERVAL); err != nil {

		log.Fatalln("error registering option circular-peer-refresh:", err)
	}

	if err := p.RegisterNewIntOption("circular-liquidity-refresh",
		"The period of time after which the liquidity is reset (minutes)",
		node.DEFAULT_LIQUIDITY_RESET_INTERVAL); err != nil {

		log.Fatalln("error registering option circular-liquidity-reset:", err)
	}
}
