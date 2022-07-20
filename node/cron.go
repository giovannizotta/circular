package node

import (
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"github.com/robfig/cron/v3"
	"log"
	"time"
)

const (
	STATS_REFRESH_INTERVAL = "10m"
)

func (n *Node) setupCronJobs(options map[string]glightning.Option) {
	c := cron.New()
	addCronJob(c, options["graph_refresh"].GetValue().(string), func() {
		n.refreshGraph()
	})
	addCronJob(c, options["peer_refresh"].GetValue().(string), func() {
		n.refreshPeers()
	})
	addCronJob(c, STATS_REFRESH_INTERVAL, func() {
		n.PrintStats()
	})
	c.Start()
}

func addCronJob(c *cron.Cron, interval string, f func()) {
	_, err := c.AddFunc("@every "+interval, f)
	if err != nil {
		log.Fatalln("error adding cron job", err)
	}
}

func (n *Node) refreshGraph() {
	defer util.TimeTrack(time.Now(), "node.refreshGraph")

	channelList, err := n.lightning.ListChannels()
	if err != nil {
		log.Printf("error listing channels: %v\n", err)
	}

	n.Graph.RefreshChannels(channelList)

	nodes, err := n.lightning.ListNodes()
	if err != nil {
		log.Printf("error listing nodes: %v\n", err)
	}
	n.Graph.RefreshAliases(nodes)
	n.Graph.SaveToFile(CIRCULAR_DIR, "graph.json")
}

func (n *Node) refreshPeers() {
	defer util.TimeTrack(time.Now(), "node.refreshPeers")
	newPeers := make(map[string]*glightning.Peer)
	peers, err := n.lightning.ListPeers()
	if err != nil {
		log.Fatalln(err)
	}
	for _, peer := range peers {
		newPeers[peer.Id] = peer
	}
	n.Peers = newPeers
}
