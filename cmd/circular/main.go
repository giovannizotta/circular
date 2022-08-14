package main

import (
	"circular/node"
	"fmt"
	"github.com/elementsproject/glightning/glightning"
	"github.com/virtuald/go-paniclog"
	"log"
	"os"
	"time"
)

var (
	lightning *glightning.Lightning
	plugin    *glightning.Plugin
)

// This is called after the plugin starts up successfully
func onInit(plugin *glightning.Plugin, options map[string]glightning.Option, config *glightning.Config) {
	circularDir := config.LightningDir + "/" + node.CIRCULAR_DIR
	// check if dir exists, otherwise create it
	if _, err := os.Stat(circularDir); os.IsNotExist(err) {
		if err := os.Mkdir(circularDir, 0755); err != nil {
			log.Fatalln(err)
		}
	}

	// we redirect stderr to a file, so that we can have panic logs logged in a file
	if err := redirectStderr(circularDir); err != nil {
		log.Fatalln(err)
	}

	lightning = glightning.NewLightning()
	if err := lightning.StartUp(config.RpcFile, config.LightningDir); err != nil {
		log.Fatalln("error starting plugin: ", err)
	}

	node.GetNode().Init(lightning, plugin, options, config)
	log.Printf("circular successfully init'd!\n")
}

func main() {
	plugin = glightning.NewPlugin(onInit)
	registerOptions(plugin)
	registerMethods(plugin)
	registerSubscriptions(plugin)
	registerHooks(plugin)

	err := plugin.Start(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalln(err)
	}
}

func redirectStderr(dir string) error {
	filename := dir + "/" + fmt.Sprintf("stderr-%d.log", time.Now().Unix())
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	_, err = paniclog.RedirectStderr(f)
	if err != nil {
		return err
	}
	return nil
}
