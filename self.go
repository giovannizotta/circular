package main

import (
	"github.com/elementsproject/glightning/glightning"
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

func RefreshPeers() map[string]*glightning.Peer {
	log.Println("refreshing peers")
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

func getId() string {
	info, err := lightning.GetInfo()
	if err != nil {
		log.Fatalln(err)
	}
	return info.Id
}
