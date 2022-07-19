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
	SENDPAY_TIMEOUT = 120 // 2 minutes
	CIRCULAR_DIR    = "circular"
	PEER_REFRESH    = "1m"
)

var (
	singleton *Node
	once      sync.Once
)

type Node struct {
	lightning *glightning.Lightning
	Id        string
	Peers     map[string]*glightning.Peer
	Graph     *graph.Graph
	DB        *Store
}

func GetNode() *Node {
	once.Do(func() {
		rand.Seed(time.Now().UnixNano())
		singleton = &Node{
			Peers: make(map[string]*glightning.Peer),
		}
	})
	return singleton
}

func (s *Node) Init(lightning *glightning.Lightning, options map[string]glightning.Option, config *glightning.Config) {
	s.lightning = lightning

	info, err := s.lightning.GetInfo()
	if err != nil {
		log.Fatalln(err)
	}
	s.Id = info.Id
	s.Graph = graph.NewGraph()
	g := graph.LoadFromFile(CIRCULAR_DIR + "/" + graph.FILE)
	if g != nil {
		s.Graph = g
	}
	s.refreshGraph()
	s.refreshPeers()
	s.DB = NewDB(config.LightningDir + "/" + CIRCULAR_DIR)
	s.setupCronJobs(options)
}

func (s *Node) PrintStats() {
	log.Println("Node stats:")
	log.Println("  Peers:", len(s.Peers))

	s.Graph.PrintStats()
	
	successes, err := s.DB.ListSuccesses()
	if err != nil {
		log.Println(err)
	}
	log.Println("successes:", len(successes))
	failures, err := s.DB.ListFailures()
	if err != nil {
		log.Println(err)
	}
	log.Println("failures:", len(failures))
}
