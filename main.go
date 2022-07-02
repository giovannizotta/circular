package main

import (
	"github.com/elementsproject/glightning/glightning"
	"github.com/robfig/cron/v3"
	"log"
	"os"
)

var (
	lightning *glightning.Lightning
	plugin    *glightning.Plugin
	graph     *Graph
	self      *Self
)

// This is called after the plugin starts up successfully
func onInit(_ *glightning.Plugin, options map[string]glightning.Option, config *glightning.Config) {
	log.Printf("successfully init'd! %s\n", config.RpcFile)
	lightning = glightning.NewLightning()
	lightning.StartUp(config.RpcFile, config.LightningDir)

	self = NewSelf()
	log.Printf("Is this initial node startup? %v\n", config.Startup)

	c := cron.New()
	SetRecurrentGraphRefresh(c, options["graph_refresh"].GetValue().(string))
	c.Start()
}

func main() {
	plugin = glightning.NewPlugin(onInit)
	registerOptions(plugin)
	registerSubscriptions(plugin)
	registerMethods(plugin)

	err := plugin.Start(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalln(err)
	}
}

func registerSubscriptions(p *glightning.Plugin) {
	//tmp
	p.SubscribeChannelOpened(OnChannelOpened)
}

func OnChannelOpened(co *glightning.ChannelOpened) {
	//tmp
	log.Printf("channel opened with %s for %s. is locked? %v", co.PeerId, co.FundingSatoshis, co.FundingLocked)
}

func registerOptions(p *glightning.Plugin) {
	err := p.RegisterNewOption("graph_refresh",
		"How often the gossip graph gets refreshed",
		GRAPH_REFRESH)
	if err != nil {
		log.Fatalln("error registering option:", err)
	}
}

func registerMethods(p *glightning.Plugin) {
	rpcRebalance := glightning.NewRpcMethod(&Rebalance{}, "Rebalance")
	rpcRebalance.LongDesc = "Rebalance the channel `In` from the channel `Out` for amount `Amount`"
	rpcRebalance.Category = "utility"
	p.RegisterMethod(rpcRebalance)
}
