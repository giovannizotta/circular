package node

import (
	"circular/graph"
	"circular/util"
	"fmt"
	"github.com/elementsproject/glightning/glightning"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	CIRCULAR_DIR = "circular"
	PEER_REFRESH = "30s"
)

var (
	singleton *Node
	once      sync.Once
)

type Node struct {
	lightning           *glightning.Lightning
	plugin              *glightning.Plugin
	initLock            *sync.Cond
	Id                  string
	Peers               map[string]*glightning.Peer
	Graph               *graph.Graph
	DB                  *Store
	LiquidityUpdateChan chan *LiquidityUpdate
}

func GetNode() *Node {
	once.Do(func() {
		rand.Seed(time.Now().UnixNano())
		singleton = &Node{
			initLock:            sync.NewCond(&sync.Mutex{}),
			Peers:               make(map[string]*glightning.Peer),
			LiquidityUpdateChan: make(chan *LiquidityUpdate, 16),
		}
		go singleton.UpdateLiquidity()
	})
	// This makes sure the node is not used until it is initialized or refreshed
	singleton.initLock.L.Lock()
	// unlock it right away.
	singleton.initLock.L.Unlock()
	return singleton
}

func (n *Node) Init(lightning *glightning.Lightning, plugin *glightning.Plugin, options map[string]glightning.Option, config *glightning.Config) {
	defer util.TimeTrack(time.Now(), "node.Init", n.Logf)
	n.lightning = lightning
	n.plugin = plugin
	n.initLock.L.Lock()
	defer n.initLock.L.Unlock()

	n.Logln(glightning.Info, "initializing node")
	n.Logln(glightning.Debug, "getting ID")
	info, err := n.lightning.GetInfo()
	if err != nil {
		log.Fatalln(err)
	}
	n.Id = info.Id

	n.Logln(glightning.Debug, "loading from file")
	err = n.LoadGraphFromFile(config.LightningDir+"/"+CIRCULAR_DIR, graph.FILE)
	if err == util.ErrNoGraphToLoad {
		// If we don't have a graph, we need to create one
		n.Logln(glightning.Unusual, err)
		n.Graph = graph.NewGraph()
	} else if err != nil {
		// If we have an error, we're in trouble
		n.Logln(glightning.Unusual, err)
		log.Fatalln(err)
	}

	n.Logln(glightning.Debug, "refreshing graph")
	err = n.refreshGraph()
	if err != nil {
		log.Fatalln(err)
	}

	n.Logln(glightning.Debug, "refreshing peers")
	err = n.refreshPeers()
	if err != nil {
		log.Fatalln(err)
	}

	n.Logln(glightning.Debug, "opening database")
	n.DB = NewDB(config.LightningDir + "/" + CIRCULAR_DIR)

	n.Logln(glightning.Debug, "setting up cronjobs")
	n.setupCronJobs(options)

	n.PrintStats()
	n.Logln(glightning.Info, "node initialized")
}

func (n *Node) Logf(level glightning.LogLevel, format string, v ...any) {
	n.plugin.Log(util.GetCallInfo()+fmt.Sprintf(format, v...), level)
}

func (n *Node) Logln(level glightning.LogLevel, v ...any) {
	n.plugin.Log(util.GetCallInfo()+fmt.Sprint(v...), level)
}

func (n *Node) PrintStats() {
	n.Logln(glightning.Info, "Node stats:")
	n.Logln(glightning.Info, "Peers:", len(n.Peers))

	n.Logln(glightning.Info, n.Graph.GetStats())

	successes, err := n.DB.ListSuccesses()
	if err != nil {
		n.Logln(glightning.Info, err)
	}
	n.Logln(glightning.Info, "successes: ", len(successes))
	var totalMoved uint64 = 0
	for _, s := range successes {
		totalMoved += s.MilliSatoshi
	}
	n.Logln(glightning.Info, "Total amount of BTC rebalanced: ", totalMoved/1000, "sats")
	failures, err := n.DB.ListFailures()
	if err != nil {
		n.Logln(glightning.Info, err)
	}
	n.Logln(glightning.Info, "failures: ", len(failures))
}

func (n *Node) RefreshChannel(channel *graph.Channel) {
	channels, err := n.lightning.GetChannel(channel.ShortChannelId)
	if err != nil {
		n.Logln(glightning.Unusual, err)
		return
	}
	n.Graph.RefreshChannels(channels)
}
