package node

import (
	"circular/graph"
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
			Peers:               make(map[string]*glightning.Peer),
			LiquidityUpdateChan: make(chan *LiquidityUpdate, 16),
		}
		go singleton.UpdateLiquidity()
	})
	return singleton
}

func (n *Node) Init(lightning *glightning.Lightning, options map[string]glightning.Option, config *glightning.Config) {
	n.lightning = lightning

	info, err := n.lightning.GetInfo()
	if err != nil {
		log.Fatalln(err)
	}
	n.Id = info.Id
	n.Graph = graph.NewGraph()
	g := graph.LoadFromFile(CIRCULAR_DIR + "/" + graph.FILE)
	if g != nil {
		n.Graph = g
	}
	n.refreshGraph()
	n.refreshPeers()
	n.DB = NewDB(config.LightningDir + "/" + CIRCULAR_DIR)
	n.setupCronJobs(options)
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
