package main

import "github.com/elementsproject/glightning/glightning"

func registerMethods(p *glightning.Plugin) {
	rpcRebalance := glightning.NewRpcMethod(&Rebalance{}, "Rebalance")
	rpcRebalance.LongDesc = "Rebalance the channel `In` from the channel `Out` for amount `Amount` for at most `MaxPPM`"
	rpcRebalance.Category = "utility"
	p.RegisterMethod(rpcRebalance)
}
