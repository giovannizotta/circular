package main

import (
	"circular/node"
	"github.com/elementsproject/glightning/glightning"
	"log"
	"math/rand"
	"os"
	"time"
)

var (
	lightning *glightning.Lightning
	plugin    *glightning.Plugin
	self      *node.Self
)

// This is called after the plugin starts up successfully
func onInit(_ *glightning.Plugin, options map[string]glightning.Option, config *glightning.Config) {
	log.Printf("successfully init'd! %s\n", config.RpcFile)
	lightning = glightning.NewLightning()
	err := lightning.StartUp(config.RpcFile, config.LightningDir)
	if err != nil {
		log.Fatalln("error starting lightning: ", err)
	}

	self = node.GetSelf()
	self.Init(lightning, options)
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
