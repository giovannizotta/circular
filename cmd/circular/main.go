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
	log.Printf("successfully init'd! %s\n", config.RpcFile)
	lightning = glightning.NewLightning()
	err := lightning.StartUp(config.RpcFile, config.LightningDir)
	if err != nil {
		log.Fatalln("error starting plugin: ", err)
	}

	node.GetNode().Init(lightning, options, config)
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
