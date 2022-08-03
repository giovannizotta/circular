package node

import (
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"github.com/robfig/cron/v3"
	"log"
	"strconv"
	"time"
)

const (
	LIQUIDITY_REFRESH_INTERVAL = 10
	STATS_REFRESH_INTERVAL     = 10
)

func (n *Node) setupCronJobs(options map[string]glightning.Option) {
	c := cron.New()

	addCronJob(c, strconv.Itoa(options["circular-graph-refresh"].GetValue().(int))+"m", func() {
		n.refreshGraph()
	})

	addCronJob(c, strconv.Itoa(options["circular-peer-refresh"].GetValue().(int))+"s", func() {
		n.refreshPeers()
	})

	addCronJob(c, strconv.Itoa(LIQUIDITY_REFRESH_INTERVAL)+"m", func() {
		n.refreshLiquidity()
	})

	addCronJob(c, strconv.Itoa(STATS_REFRESH_INTERVAL)+"m", func() {
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
	defer util.TimeTrack(time.Now(), "node.refreshGraph", n.Logf)

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
	err = n.SaveGraphToFile(CIRCULAR_DIR, "graph.json")
	if err != nil {
		n.Logf(glightning.Unusual, "error saving graph to file: %v\n", err)
		return err
	}
	return nil
}

func (n *Node) refreshPeers() error {
	defer util.TimeTrack(time.Now(), "node.refreshPeers", n.Logf)

	n.Logln(glightning.Debug, "refreshing peers")
	peers, err := n.lightning.ListPeers()
	if err != nil {
		n.Logln(glightning.Unusual, err)
		return err
	}

	n.PeersLock.Lock()
	defer n.PeersLock.Unlock()
	for _, peer := range peers {
		n.Peers[peer.Id] = peer
	}
	return nil
}
