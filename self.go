package main

import (
	"github.com/elementsproject/glightning/glightning"
	"github.com/robfig/cron/v3"
	"log"
)

const (
	PEER_REFRESH = "10s"
)

type Self struct {
	Id    string
	Peers map[string]*glightning.Peer
}

func NewSelf() *Self {
	result := &Self{}
	result.Id = getId()
	result.Peers = make(map[string]*glightning.Peer)
	return result
}

func refreshPeers() map[string]*glightning.Peer {
	newPeers := make(map[string]*glightning.Peer)
	peers, err := lightning.ListPeers()
	if err != nil {
		log.Fatalln(err)
	}
	for _, peer := range peers {
		newPeers[peer.Id] = peer
	}
	return newPeers
}

func (s *Self) SetRecurrentPeersRefresh(c *cron.Cron, peerRefresh string) {
	_, err := c.AddFunc("@every "+peerRefresh, func() {
		s.Peers = refreshPeers()
	})
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func getId() string {
	info, err := lightning.GetInfo()
	if err != nil {
		log.Fatalln(err)
	}
	return info.Id
}
