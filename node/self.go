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
	PEER_REFRESH = "1m"
)

var (
	singleton *Self

	once sync.Once
)

type Self struct {
	lightning *glightning.Lightning
	Id        string
	Peers     map[string]*glightning.Peer
	Graph     *graph.Graph
	DB        *PreimageStore
}

func GetSelf() *Self {
	once.Do(func() {
		rand.Seed(time.Now().UnixNano())
		singleton = &Self{
			Peers: make(map[string]*glightning.Peer),
		}
	})
	return singleton
}

func (s *Self) Init(lightning *glightning.Lightning, options map[string]glightning.Option, config *glightning.Config) {
	s.lightning = lightning

	info, err := s.lightning.GetInfo()
	if err != nil {
		log.Fatalln(err)
	}
	s.Id = info.Id
	s.Graph = graph.NewGraph()
	s.setupCronJobs(options)
	s.DB = NewDB(config.LightningDir)
}

func (s *Self) refreshPeers() {
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

func (s *Self) refreshGraph() {
	newGraph := &graph.Graph{}

	channelList, err := s.lightning.ListChannels()
	if err != nil {
		log.Printf("error listing channels: %v\n", err)
	}

	for _, c := range channelList {
		newGraph.AddChannel(c)
	}

	s.Graph.Inbound = newGraph.Inbound
	s.Graph.Outbound = newGraph.Outbound
}

func (s *Self) GetBestPeerChannel(id string, metric func(*glightning.PeerChannel) uint64) *glightning.PeerChannel {
	channels := s.Peers[id].Channels
	best := channels[0]
	for _, channel := range channels {
		if metric(channel) > metric(best) {
			best = channel
		}
	}
	return best
}

func (s *Self) HasPeer(id string) bool {
	return s.Peers[id] != nil
}

func addCronJob(c *cron.Cron, interval string, f func()) {
	_, err := c.AddFunc("@every "+interval, f)
	if err != nil {
		log.Fatalln("error adding cron job", err)
	}
}

func (s *Self) setupCronJobs(options map[string]glightning.Option) {
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

func (s *Self) SendPay(route *graph.Route, paymentHash string) (*glightning.SendPayFields, error) {
	_, err := s.lightning.SendPayLite(route.ToLightningRoute(), paymentHash)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// TODO: learn from failed payments
	result, err := s.lightning.WaitSendPay(paymentHash, 20)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Self) GeneratePreimageHashPair() (string, error) {
	pair := NewPreimageHashPair()
	err := s.DB.Set(pair.Hash, pair.Preimage)
	if err != nil {
		return "", err
	}
	return pair.Hash, nil
}
