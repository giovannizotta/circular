package node

import (
	"circular/graph"
	"circular/util"
	"github.com/dgraph-io/badger/v3"
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
	CIRCULAR_DIR    = "circular"
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
	g := graph.LoadFromFile(CIRCULAR_DIR + "/" + graph.FILE)
	if g != nil {
		s.Graph = g
	}
	s.refreshGraph()
	s.refreshPeers()
	s.DB = NewDB(config.LightningDir + "/" + CIRCULAR_DIR)
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

	result, err := s.lightning.WaitSendPay(paymentHash, PAYMENT_TIMEOUT)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return result, nil
}

func (s *Node) deleteIfOurs(paymentHash string) error {
	_, err := s.DB.Get(paymentHash)
	if err == badger.ErrKeyNotFound {
		return err // this payment was not made by us
	}
	err = s.DB.Delete(paymentHash)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (s *Node) OnPaymentFailure(sf *glightning.SendPayFailure) {
	if err := s.deleteIfOurs(sf.Data.PaymentHash); err != nil {
		return
	}
	direction := strconv.Itoa(sf.Data.ErringDirection)
	oppositeDirection := strconv.Itoa(sf.Data.ErringDirection ^ 0x1)
	channelId := sf.Data.ErringChannel + "/" + direction
	oppositeChannelId := sf.Data.ErringChannel + "/" + oppositeDirection
	log.Println("failed from " + s.Graph.Channels[oppositeChannelId].Source + " to " + s.Graph.Channels[oppositeChannelId].Destination)
	log.Printf("channel %s failed, opposite channel is %s\n", oppositeChannelId, channelId)
	log.Printf("code: %d, failcode: %d, failcodename: %s\n", sf.Code, sf.Data.FailCode, sf.Data.FailCodeName)
	// TODO: handle failure codes separately: right now we treat every failure as a liquidity failure, but it might not be the case
	s.Graph.Channels[oppositeChannelId].Liquidity = sf.Data.MilliSatoshi - 1000000
	s.Graph.Channels[channelId].Liquidity =
		s.Graph.Channels[channelId].Satoshis*1000 - s.Graph.Channels[oppositeChannelId].Liquidity
}

func (s *Node) OnPaymentSuccess(ss *glightning.SendPaySuccess) {
	s.deleteIfOurs(ss.PaymentHash)
}
