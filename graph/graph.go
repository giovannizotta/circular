package graph

import (
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	GRAPH_REFRESH             = "10m"
	FILE                      = "graph.json"
	PRUNING_INTERVAL     uint = 1209600 // 14 days
	AVERAGE_AGING_AMOUNT      = 0       // the amount by which the liquidity belief is updated
	AGING_VARIANCE            = 0       // the range (+/-) of the random amount added to the liquidity belief
	// for example, for an AVERAGE_AGING_AMOUNT of 10k and an AGING_VARIANCE of 5k
	// the liquidity belief will be updated by a random amount between 5k and 15k	(10k +- 5k)
)

// Edge contains All the SCIDs of the channels going from nodeA to nodeB
type Edge []string

// Graph is the lightning network graph from the perspective of our node
// It has been built from the gossip received by lightningd.
// To access the edges flowing into a node, use: g.Inbound[node]
// To access an edge into nodeA from nodeB, use: g.Inbound[nodeA][nodeB]
// * an edge consists of one or more SCIDs between nodeA and nodeB
// To access a channel via channelId (scid/direction). use: g.Channels[channelId]
type Graph struct {
	Channels map[string]*Channel        `json:"channels"`
	Inbound  map[string]map[string]Edge `json:"-"`
	Aliases  map[string]string          `json:"-"`
}

func NewGraph() *Graph {
	return &Graph{
		Channels: make(map[string]*Channel),
		Inbound:  make(map[string]map[string]Edge),
		Aliases:  make(map[string]string),
	}
}

func allocate(links *map[string]map[string]Edge, from, to string) {
	if (*links)[from] == nil {
		(*links)[from] = make(map[string]Edge)
	}
	if (*links)[from][to] == nil {
		(*links)[from][to] = make([]string, 0)
	}
}

func (g *Graph) AddChannel(c *Channel) {
	allocate(&g.Inbound, c.Destination, c.Source)
	g.Inbound[c.Destination][c.Source] = append(g.Inbound[c.Destination][c.Source], c.ShortChannelId)
}

func (g *Graph) RefreshChannels(channelList []*glightning.Channel) {
	// TODO: remove stale channels
	// we need to do NewChannel and not only update the liquidity because of gossip updates
	defer util.TimeTrack(time.Now(), "graph.RefreshChannels")
	for _, c := range channelList {
		var channel *Channel
		channelId := c.ShortChannelId + "/" + util.GetDirection(c.Source, c.Destination)
		// if the channel did not exist prior to this refresh estimate its initial liquidity to be 50/50
		if _, ok := g.Channels[channelId]; !ok {
			channel = NewChannel(c, uint64(0.5*float64(c.Satoshis*1000)))
			g.AddChannel(channel)
		} else {
			channel = NewChannel(c, g.Channels[channelId].Liquidity)
		}
		g.Channels[channelId] = channel
	}
}

func (g *Graph) getLiquidityAfterAging(channelId string, perfectBalance uint64) uint64 {
	aging := util.RandRange(AVERAGE_AGING_AMOUNT-AGING_VARIANCE, AVERAGE_AGING_AMOUNT+AGING_VARIANCE)
	return util.Min(g.Channels[channelId].Liquidity+aging, perfectBalance)
}

func (g *Graph) updateOppositeChannel(c *Channel, liquidity uint64) {
	oppositeChannelId := c.ShortChannelId + "/" + util.GetDirection(c.Destination, c.Source)
	// if opposite channel is in the map
	if _, ok := g.Channels[oppositeChannelId]; ok {
		oppositeChannel := g.Channels[oppositeChannelId]
		oppositeChannel.Liquidity = (c.Satoshis * 1000) - liquidity
	}
}

func (g *Graph) RefreshAliases(nodes []*glightning.Node) {
	for _, n := range nodes {
		g.Aliases[n.Id] = n.Alias
	}
}

func (g *Graph) PrintStats() {
	activeChannels := 0
	atLeast200kLiquidity := 0
	atLeast200kMaxHtlc := 0
	for _, c := range g.Channels {
		if c.IsActive {
			activeChannels++
		}
		if c.Liquidity >= 200000000 {
			atLeast200kLiquidity++
		}
		maxHtlc, _ := strconv.ParseUint(strings.TrimSuffix(c.HtlcMaximumMilliSatoshis, "msat"), 10, 64)
		if maxHtlc >= 200000000 {
			atLeast200kMaxHtlc++
		}
	}
	log.Println("Graph stats:")
	log.Println("graph has", len(g.Inbound), "nodes")
	log.Println("graph has", len(g.Channels), "channels")
	log.Println("graph has", activeChannels, "active channels")
	log.Println("graph has", atLeast200kLiquidity, "channels believed to have at least 200k liquidity")
	log.Println("graph has", atLeast200kMaxHtlc, "channels with at least 200k max htlc")
}

func (g *Graph) PruneChannels() {
	// get current time in seconds
	now := uint(time.Now().Unix())

	// prune channels that are older than PRUNING_INTERVAL
	// TODO: remove closed channels, but might be worth waiting for glightning to implement channel_state_changed
	for _, c := range g.Channels {
		if c.LastUpdate+PRUNING_INTERVAL < now {
			delete(g.Channels, c.ShortChannelId+"/0")
			delete(g.Channels, c.ShortChannelId+"/1")
		}
	}
}
