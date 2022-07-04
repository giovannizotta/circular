package main

import (
	"github.com/elementsproject/glightning/glightning"
	"github.com/robfig/cron/v3"
	"log"
	"math/rand"
	"os"
	"time"
)

var (
	lightning         *glightning.Lightning
	plugin            *glightning.Plugin
	graph             *Graph
	self              *Self
	ongoingRebalances map[string]*Rebalance
)

// This is called after the plugin starts up successfully
func onInit(_ *glightning.Plugin, options map[string]glightning.Option, config *glightning.Config) {
	log.Printf("successfully init'd! %s\n", config.RpcFile)
	lightning = glightning.NewLightning()
	lightning.StartUp(config.RpcFile, config.LightningDir)

	ongoingRebalances = make(map[string]*Rebalance)
	self = NewSelf()
	log.Printf("Is this initial node startup? %v\n", config.Startup)

	// TODO: refactor this to be more generic
	c := cron.New()
	SetRecurrentGraphRefresh(c, options["graph_refresh"].GetValue().(string))
	self.SetRecurrentPeersRefresh(c, options["peer_refresh"].GetValue().(string))
	c.Start()
}

func main() {
	rand.Seed(time.Now().UnixNano())
	plugin = glightning.NewPlugin(onInit)
	registerOptions(plugin)
	registerSubscriptions(plugin)
	registerMethods(plugin)
	registerHooks(plugin)

	err := plugin.Start(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalln(err)
	}
}
