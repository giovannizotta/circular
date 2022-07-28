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
		n.initLock.L.Lock()
		n.refreshGraph()
		n.initLock.L.Unlock()
	})
	addCronJob(c, options["peer_refresh"].GetValue().(string), func() {
		n.initLock.L.Lock()
		n.refreshPeers()
		n.initLock.L.Unlock()
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

func (n *Node) refreshGraph() error {
	defer n.Logln(glightning.Debug, util.TimeTrack(time.Now(), "node.refreshGraph"))
	channelList, err := n.lightning.ListChannels()
	if err != nil {
		n.Logf(glightning.Unusual, "error listing channels: %v\n", err)
		return err
	}

	n.Logln(glightning.Debug, "refreshing channels")
	n.Graph.RefreshChannels(channelList)

	n.Logln(glightning.Debug, "pruning channels")
	n.Graph.PruneChannels()

	n.Logln(glightning.Debug, "refreshing aliases")
	nodes, err := n.lightning.ListNodes()
	if err != nil {
		n.Logf(glightning.Unusual, "error listing nodes: %v\n", err)
		return err
	}
	n.Graph.RefreshAliases(nodes)

	n.Logln(glightning.Debug, "saving graph to file")
	n.Graph.SaveToFile(CIRCULAR_DIR, "graph.json")
	return nil
}

func (n *Node) refreshPeers() error {
	defer n.Logln(glightning.Debug, util.TimeTrack(time.Now(), "node.refreshPeers"))
	n.Logln(glightning.Debug, "refreshing peers")
	peers, err := n.lightning.ListPeers()
	if err != nil {
		n.Logln(glightning.Unusual, err)
		return err
	}
	for _, peer := range peers {
		n.Peers[peer.Id] = peer
	}
	return nil
}
