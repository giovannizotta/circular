package main

import (
	"circular/node"
	"circular/rebalance"
	"circular/rebalance/parallel"
	"github.com/elementsproject/glightning/glightning"
)

func registerMethods(p *glightning.Plugin) {
	rpcRebalanceByNode := glightning.NewRpcMethod(&rebalance.RebalanceByNode{}, "Rebalance by NodeID")
	rpcRebalanceByNode.LongDesc = "Rebalance the node `innode` from the node `outnode` for amount `amount` for at most `maxppm`"
	rpcRebalanceByNode.Category = "utility"
	p.RegisterMethod(rpcRebalanceByNode)

	rpcRebalanceByScid := glightning.NewRpcMethod(&rebalance.RebalanceByScid{}, "Rebalance by Scid")
	rpcRebalanceByScid.LongDesc = "Rebalance the channel `inchannel` from the channel `outchannel` for amount `amount` for at most `maxppm`"
	rpcRebalanceByScid.Category = "utility"
	p.RegisterMethod(rpcRebalanceByScid)

	rpcRebalancePull := glightning.NewRpcMethod(&parallel.RebalancePull{}, "Pull liquidity into a channel from many sources in parallel")
	rpcRebalancePull.LongDesc = "Rebalance the channel `inscid` from many channels concurrently"
	rpcRebalancePull.Category = "utility"
	p.RegisterMethod(rpcRebalancePull)

	rpcRebalancePush := glightning.NewRpcMethod(&parallel.RebalancePush{}, "Push liquidity out of a channel to many destinations in parallel")
	rpcRebalancePush.LongDesc = "Rebalance the channel `outscid` from many channels concurrently"
	rpcRebalancePush.Category = "utility"
	p.RegisterMethod(rpcRebalancePush)

	rpcStats := glightning.NewRpcMethod(&node.Stats{}, "Get stats")
	rpcStats.LongDesc = "Get the stats of the usage of circular"
	rpcStats.Category = "utility"
	p.RegisterMethod(rpcStats)

	deleteStats := glightning.NewRpcMethod(&node.DeleteStats{}, "Delete Stats")
	deleteStats.LongDesc = "Delete the stats of the usage of circular"
	deleteStats.Category = "utility"
	p.RegisterMethod(deleteStats)

}
