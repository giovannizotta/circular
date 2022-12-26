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
	LIQUIDITY_REFRESH_INTERVAL = 10 // minutes
)

func (n *Node) setupCronJobs(options map[string]glightning.Option) {
	c := cron.New()

	// every 10 minutes by default, refresh the information gathered via gossip
	addCronJob(c, strconv.Itoa(options["circular-graph-refresh"].GetValue().(int))+"m", func() {
		n.refreshGraph()
	})

	// every 30 seconds by default, refresh peers
	addCronJob(c, strconv.Itoa(options["circular-peer-refresh"].GetValue().(int))+"s", func() {
		n.refreshPeers()
	})

	// every 10 minutes by default, check if there are channels that need to be reset
	addCronJob(c, strconv.Itoa(LIQUIDITY_REFRESH_INTERVAL)+"m", func() {
		n.refreshLiquidity()
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
	n.Logln(glightning.Info, "refreshing graph")

	channelList, err := n.lightning.ListChannels()
	if err != nil {
		n.Logf(glightning.Unusual, "error listing channels: %+v", err)
		return err
	}

	n.Logln(glightning.Info, "refreshing channels")
	n.Graph.RefreshChannels(channelList)

	n.Logln(glightning.Info, "pruning channels")
	n.Graph.PruneChannels()

	n.Logln(glightning.Info, "refreshing aliases")
	nodes, err := n.lightning.ListNodes()
	if err != nil {
		n.Logf(glightning.Unusual, "error listing nodes: %+v", err)
		return err
	}
	n.Graph.RefreshAliases(nodes)

	n.Logln(glightning.Info, "saving graph to file")
	if err = n.SaveGraphToFile(CIRCULAR_DIR, "graph.json"); err != nil {
		n.Logf(glightning.Unusual, "error saving graph to file: %+v", err)
		return err
	}

	n.Logln(glightning.Info, "graph has been refreshed")
	return nil
}

func (n *Node) refreshPeers() error {
	defer util.TimeTrack(time.Now(), "node.refreshPeers", n.Logf)
	n.Logln(glightning.Info, "refreshing peers")

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

func (n *Node) refreshLiquidity() {
	defer util.TimeTrack(time.Now(), "node.refreshLiquidity", n.Logf)
	n.Logln(glightning.Info, "refreshing liquidity")

	hits := n.Graph.RefreshLiquidity(n.liquidityRefresh)
	n.Logf(glightning.Info, "liquidity has been reset on %d channels", hits)
}
