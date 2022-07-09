package node

import (
	"circular/graph"
	"github.com/elementsproject/glightning/glightning"
	"github.com/robfig/cron/v3"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	PEER_REFRESH    = "1m"
	PAYMENT_TIMEOUT = 60
)

var (
	singleton *Node

	once sync.Once
)

type Node struct {
	lightning *glightning.Lightning
	Id        string
	Peers     map[string]*glightning.Peer
	Graph     *graph.Graph
	DB        *PreimageStore
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
	if s.Graph == nil {
		log.Println("could not retrieve graph from file")
		s.refreshGraph()
	}
	s.refreshPeers()
	s.DB = NewDB(config.LightningDir)
	s.setupCronJobs(options)
}

func (s *Node) refreshPeers() {
	log.Println("refreshing peers")
	newPeers := make(map[string]*glightning.Peer)
	peers, err := s.lightning.ListPeers()
	if err != nil {
		log.Fatalln(err)
	}
	for _, peer := range peers {
		newPeers[peer.Id] = peer
	}
	s.Peers = newPeers
}

func (s *Node) refreshGraph() {
	log.Println("refreshing graph")
	newGraph := &graph.Graph{}

	channelList, err := s.lightning.ListChannels()
	if err != nil {
		log.Printf("error listing channels: %v\n", err)
	}

	for _, c := range channelList {
		newGraph.AddChannel(c)
	}
	s.Graph = newGraph
}

func (s *Node) GetBestPeerChannel(id string, metric func(*glightning.PeerChannel) uint64) *glightning.PeerChannel {
	channels := s.Peers[id].Channels
	best := channels[0]
	for _, channel := range channels {
		if metric(channel) > metric(best) {
			best = channel
		}
	}
	return best
}

func (s *Node) HasPeer(id string) bool {
	return s.Peers[id] != nil
}

func addCronJob(c *cron.Cron, interval string, f func()) {
	_, err := c.AddFunc("@every "+interval, f)
	if err != nil {
		log.Fatalln("error adding cron job", err)
	}
}

func (s *Node) setupCronJobs(options map[string]glightning.Option) {
	c := cron.New()
	addCronJob(c, options["graph_refresh"].GetValue().(string), func() {
		s.refreshGraph()
		s.Graph.SaveToFile()
	})
	addCronJob(c, options["peer_refresh"].GetValue().(string), func() {
		s.refreshPeers()
	})
	c.Start()
}

func (s *Node) SendPay(route *graph.Route, paymentHash string) (*glightning.SendPayFields, error) {
	_, err := s.lightning.SendPayLite(route.ToLightningRoute(), paymentHash)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// TODO: learn from failed payments
	result, err := s.lightning.WaitSendPay(paymentHash, PAYMENT_TIMEOUT)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Node) GeneratePreimageHashPair() (string, error) {
	pair := NewPreimageHashPair()
	err := s.DB.Set(pair.Hash, pair.Preimage)
	if err != nil {
		return "", err
	}
	return pair.Hash, nil
}
