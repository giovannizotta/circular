package main

import "github.com/elementsproject/glightning/glightning"

func registerMethods(p *glightning.Plugin) {
	rpcRebalance := glightning.NewRpcMethod(&Rebalance{}, "Rebalance")
	rpcRebalance.LongDesc = "Rebalance the channel `Destination` from the channel `Source` for amount `Amount` for at most `MaxPPM`"
	rpcRebalance.Category = "utility"
	p.RegisterMethod(rpcRebalance)
}
