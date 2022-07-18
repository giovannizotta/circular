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
	PEER_REFRESH    = "1m"
	SENDPAY_TIMEOUT = 120 // 2 minutes
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

func (s *Node) GetPeerChannelFromNodeID(scid string) (*glightning.PeerChannel, error) {
	for _, peer := range s.Peers {
		for _, channel := range peer.Channels {
			if channel.ShortChannelId == scid {
				return channel, nil
			}
		}
	}
	return nil, util.ErrNoPeerChannel
}

func (s *Node) HasPeer(id string) bool {
	_, ok := s.Peers[id]
	return ok
}

func (s *Node) GetChannelPeerFromScid(scid string) (*glightning.Peer, error) {
	for _, peer := range s.Peers {
		for _, channel := range peer.Channels {
			if channel.ShortChannelId == scid {
				return peer, nil
			}
		}
	}
	return nil, util.ErrNoPeerChannel
}
