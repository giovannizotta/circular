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

func addCronJob(c *cron.Cron, interval string, f func()) {
	_, err := c.AddFunc("@every "+interval, f)
	if err != nil {
		log.Fatalln("error adding cron job", err)
	}
}

func setupCronJobs(options map[string]glightning.Option) {
	c := cron.New()
	addCronJob(c, options["graph_refresh"].GetValue().(string), func() {
		graph = RefreshGraph()
	})
	addCronJob(c, options["peer_refresh"].GetValue().(string), func() {
		self.Peers = RefreshPeers()
	})
	c.Start()
}

// This is called after the plugin starts up successfully
func onInit(_ *glightning.Plugin, options map[string]glightning.Option, config *glightning.Config) {
	log.Printf("successfully init'd! %s\n", config.RpcFile)
	lightning = glightning.NewLightning()
	err := lightning.StartUp(config.RpcFile, config.LightningDir)
	if err != nil {
		log.Fatalln("error starting lightning: ", err)
	}

	ongoingRebalances = make(map[string]*Rebalance)
	self = NewSelf()

	setupCronJobs(options)
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
