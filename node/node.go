package node

import (
	"circular/graph"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	CIRCULAR_DIR = "circular"
	PEER_REFRESH = "1m"
)

var (
	singleton *Node
	once      sync.Once
)

type Node struct {
	lightning           *glightning.Lightning
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

func (n *Node) Init(lightning *glightning.Lightning, options map[string]glightning.Option, config *glightning.Config) {
	defer util.TimeTrack(time.Now(), "Node.Init")
	n.lightning = lightning
	n.initLock.L.Lock()
	defer n.initLock.L.Unlock()

	info, err := n.lightning.GetInfo()
	if err != nil {
		log.Fatalln(err)
	}
	n.Id = info.Id

	g, err := graph.LoadFromFile(config.LightningDir + "/" + CIRCULAR_DIR + "/" + graph.FILE)
	if err == nil {
		// If we have a graph, we're good to go
		n.Graph = g
	} else if err == util.ErrNoGraphToLoad {
		// If we don't have a graph, we need to create one
		log.Println(err)
		n.Graph = graph.NewGraph()
	} else {
		// If we have an error, we're in trouble
		log.Fatalln(err)
	}

	n.refreshGraph()
	n.refreshPeers()
	n.DB = NewDB(config.LightningDir + "/" + CIRCULAR_DIR)
	n.setupCronJobs(options)
	n.PrintStats()
}

func (n *Node) PrintStats() {
	log.Println("Node stats:")
	log.Println("  Peers:", len(n.Peers))

	n.Graph.PrintStats()

	successes, err := n.DB.ListSuccesses()
	if err != nil {
		log.Println(err)
	}
	log.Println("successes:", len(successes))
	failures, err := n.DB.ListFailures()
	if err != nil {
		log.Println(err)
	}
	log.Println("failures:", len(failures))
}
