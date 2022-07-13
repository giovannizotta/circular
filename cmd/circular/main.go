package main

import (
	"circular/node"
	"github.com/elementsproject/glightning/glightning"
	"log"
	"os"
)

var (
	lightning *glightning.Lightning
	plugin    *glightning.Plugin
)

// This is called after the plugin starts up successfully
func onInit(_ *glightning.Plugin, options map[string]glightning.Option, config *glightning.Config) {
	lightning = glightning.NewLightning()
	err := lightning.StartUp(config.RpcFile, config.LightningDir)
	if err != nil {
		log.Fatalln("error starting plugin: ", err)
	}

	node.GetNode().Init(lightning, options, config)
	log.Printf("circular successfully init'd!\n")
}

func main() {
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
