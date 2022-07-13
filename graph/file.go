package graph

import (
	"circular/util"
	"encoding/json"
	"log"
	"os"
	"time"
)

func LoadFromFile(filename string) *Graph {
	defer util.TimeTrack(time.Now(), "graph.LoadFromFile")
	file, err := os.Open(filename)
	if err != nil {
		log.Println("unable to load from file:", err)
		log.Println("trying to load an old version of the graph")
		file, err = os.Open(filename + ".old")
		if err != nil {
			log.Println("unable to load any old version of the file: ", err)
			return nil
		}
	}
	defer file.Close()
	g := NewGraph()
	err = json.NewDecoder(file).Decode(&g)
	if err != nil {
		return nil
	}

	for _, c := range g.Channels {
		g.AddChannel(c)
	}
	return g
}

func (g *Graph) SaveToFile(dir, filename string) {
	defer util.TimeTrack(time.Now(), "graph.SaveToFile")
	//check if dir exists, otherwise create it
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			log.Println("unable to create directory:", err)
		}
	}
	// open temporary file
	filename = dir + "/" + filename
	file, err := os.Create(filename + ".tmp")
	if err != nil {
		log.Printf("error opening file: %v", err)
		return
	}
	defer file.Close()
	// write json
	err = json.NewEncoder(file).Encode(g)
	if err != nil {
		log.Printf("error writing file: %v", err)
		return
	}

	// save old file
	// check if filename exists
	if _, err := os.Stat(filename); err == nil {
		err = os.Rename(filename, filename+".old")
	}
	// rename tmp to filename
	err = os.Rename(filename+".tmp", filename)
}
