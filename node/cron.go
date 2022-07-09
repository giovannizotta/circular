package node

import (
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

func (s *Node) setupCronJobs(options map[string]glightning.Option) {
	c := cron.New()
	addCronJob(c, options["graph_refresh"].GetValue().(string), func() {
		s.refreshGraph()
	})
	addCronJob(c, options["peer_refresh"].GetValue().(string), func() {
		s.refreshPeers()
	})
	c.Start()
}

func addCronJob(c *cron.Cron, interval string, f func()) {
	_, err := c.AddFunc("@every "+interval, f)
	if err != nil {
		log.Fatalln("error adding cron job", err)
	}
}

func (s *Node) refreshGraph() {
	defer util.TimeTrack(time.Now(), "node.refreshGraph")
	log.Println("refreshing graph")

	channelList, err := s.lightning.ListChannels()
	if err != nil {
		log.Printf("error listing channels: %v\n", err)
	}

	s.Graph.Refresh(channelList)
	s.Graph.SaveToFile()
}

func (s *Node) refreshPeers() {
	defer util.TimeTrack(time.Now(), "node.refreshPeers")
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
