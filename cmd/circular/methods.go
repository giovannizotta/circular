package main

import (
	"circular/rebalance"
	"github.com/elementsproject/glightning/glightning"
)

func registerMethods(p *glightning.Plugin) {
	rpcRebalance := glightning.NewRpcMethod(&rebalance.Rebalance{}, "Rebalance")
	rpcRebalance.LongDesc = "Rebalance the node `Destination` from the node `Source` for amount `Amount` for at most `MaxPPM`"
	rpcRebalance.Category = "utility"
	p.RegisterMethod(rpcRebalance)
}
