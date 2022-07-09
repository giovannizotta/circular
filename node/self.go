package node

import (
	"circular/graph"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"log"
	"math/rand"
	"strconv"
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
	s.refreshGraph()
	s.refreshPeers()
	s.DB = NewDB(config.LightningDir)
	s.setupCronJobs(options)
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
	_, ok := s.Peers[id]
	return ok
}

func (s *Node) SendPay(route *graph.Route, paymentHash string) (*glightning.SendPayFields, error) {
	defer util.TimeTrack(time.Now(), "node.SendPay")
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

func (s *Node) OnPaymentFailure(sf *glightning.SendPayFailure) {
	direction := strconv.Itoa(sf.Data.ErringDirection)
	oppositeDirection := strconv.Itoa(sf.Data.ErringDirection ^ 0x1)
	channelId := sf.Data.ErringChannel + "/" + direction
	oppositeChannelId := sf.Data.ErringChannel + "/" + oppositeDirection

	s.Graph.Channels[channelId].Liquidity = sf.Data.MilliSatoshi - 1000000
	s.Graph.Channels[oppositeChannelId].Liquidity =
		s.Graph.Channels[oppositeChannelId].Satoshis*1000 - s.Graph.Channels[channelId].Liquidity
}
