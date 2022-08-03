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
	CIRCULAR_DIR                     = "circular"
	DEFAULT_PEER_REFRESH_INTERVAL    = 30  // seconds
	DEFAULT_LIQUIDITY_RESET_INTERVAL = 300 // minutes
)

var (
	singleton *Node
	once      sync.Once
)

type Node struct {
	lightning           *glightning.Lightning
	plugin              *glightning.Plugin
	liquidityRefresh    time.Duration
	initLock            *sync.Mutex
	PeersLock           *sync.RWMutex
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
			initLock:            &sync.Mutex{},
			PeersLock:           &sync.RWMutex{},
			Peers:               make(map[string]*glightning.Peer),
			LiquidityUpdateChan: make(chan *LiquidityUpdate, 16),
		}
		go singleton.UpdateLiquidity()
	})
	// This makes sure the node is not used until it is initialized or refreshed
	singleton.initLock.Lock()
	// unlock it right away.
	singleton.initLock.Unlock()
	return singleton
}

func (n *Node) Init(lightning *glightning.Lightning, plugin *glightning.Plugin, options map[string]glightning.Option, config *glightning.Config) {
	defer util.TimeTrack(time.Now(), "node.Init", n.Logf)
	n.initLock.Lock()
	defer n.initLock.Unlock()

	n.lightning = lightning
	n.plugin = plugin
	n.liquidityRefresh = time.Duration(options["circular-liquidity-refresh"].GetValue().(int)) * time.Minute
	n.Logln(glightning.Debug, "liquidity refresh interval: ", int(n.liquidityRefresh.Minutes()), " minutes")

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

	n.Logln(glightning.Debug, n.GetStats().String())
	n.Logln(glightning.Info, "node initialized")
}

func (n *Node) Logf(level glightning.LogLevel, format string, v ...any) {
	n.plugin.Log(util.GetCallInfo()+fmt.Sprintf(format, v...), level)
}

func (n *Node) Logln(level glightning.LogLevel, v ...any) {
	n.plugin.Log(util.GetCallInfo()+fmt.Sprint(v...), level)
}

func (n *Node) RefreshChannel(channel *graph.Channel) {
	channels, err := n.lightning.GetChannel(channel.ShortChannelId)
	if err != nil {
		n.Logln(glightning.Unusual, err)
		return
	}
	n.Graph.RefreshChannels(channels)
}

func (n *Node) OnConnect(c *glightning.ConnectEvent) {
	n.PeersLock.Lock()
	defer n.PeersLock.Unlock()

	if _, ok := n.Peers[c.PeerId]; ok {
		n.Peers[c.PeerId].Connected = true
	}
}

func (n *Node) OnDisconnect(c *glightning.DisconnectEvent) {
	n.PeersLock.Lock()
	defer n.PeersLock.Unlock()

	if _, ok := n.Peers[c.PeerId]; ok {
		n.Peers[c.PeerId].Connected = false
	}
}
