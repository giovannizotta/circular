package node

import (
	"circular/graph"
	"circular/util"
	"encoding/json"
	"github.com/elementsproject/glightning/glightning"
	"os"
	"time"
)

func (n *Node) LoadGraphFromFile(dir, filename string) error {
	defer util.TimeTrack(time.Now(), "graph.LoadGraphFromFile", n.Logf)
	file, err := os.Open(dir + "/" + filename)
	if err != nil {
		n.Logln(glightning.Debug, "unable to load graph data:", err, ", looking for an old file")
		n.Logln(glightning.Debug, "trying to load an old version of the graph")
		filename += ".old"
		file, err = os.Open(dir + "/" + filename)
		if err != nil {
			n.Logln(glightning.Debug, "unable to load any old version of the graph: ", err, ", continuing with a new graph")
			return util.ErrNoGraphToLoad
		}
	}
	defer file.Close()
	n.Logln(glightning.Debug, "loading graph data from file:", dir+"/"+filename)

	g := graph.NewGraph()

	err = json.NewDecoder(file).Decode(g)
	if err != nil {
		return err
	}

	for _, c := range g.Channels {
		g.AddChannel(c)
	}

	n.Graph = g

	n.Logln(glightning.Info, "graph loaded successfully")
	return nil
}

func (n *Node) SaveGraphToFile(dir, filename string) error {
	defer util.TimeTrack(time.Now(), "graph.SaveGraphToFile", n.Logf)

	// check if dir exists, otherwise create it
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}
	}

	filename = dir + "/" + filename
	if err := n.serializeToFile(filename); err != nil {
		return err
	}

	// save old file
	// check if filename exists
	if _, err := os.Stat(filename); err == nil {
		err = os.Rename(filename, filename+".old")
	}
	// rename tmp to filename
	if err := os.Rename(filename+".tmp", filename); err != nil {
		return err
	}

	return nil
}

func (n *Node) serializeToFile(filename string) error {
	// open temporary file
	file, err := os.Create(filename + ".tmp")
	if err != nil {
		return err
	}
	defer file.Close()

	// write json
	n.Graph.Lock()
	defer n.Graph.Unlock()
	if err := json.NewEncoder(file).Encode(n.Graph); err != nil {
		return err
	}
	return nil
}
