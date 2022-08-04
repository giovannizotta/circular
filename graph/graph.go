package graph

import (
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"sync"
	"time"
)

const (
	FILE                                = "graph.json"
	DEFAULT_GRAPH_REFRESH_INTERVAL      = 10      // minutes
	PRUNING_INTERVAL               uint = 1209600 // 14 days
)

// Edge contains All the SCIDs of the channels going from nodeA to nodeB
type Edge []string

// Graph is the lightning network graph from the perspective of our node
// It has been built from the gossip received by lightningd.
// To access the edges flowing into a node, use: g.Inbound[node]
// To access an edge into nodeA from nodeB, use: g.Inbound[nodeA][nodeB]
// * an edge consists of an array of SCIDs between nodeA and nodeB
// To access a channel via channelId (scid/direction). use: g.Channels[channelId]
type Graph struct {
	Channels          map[string]*Channel        `json:"channels"`
	Inbound           map[string]map[string]Edge `json:"-"`
	Aliases           map[string]string          `json:"-"`
	adjacencyListLock *sync.RWMutex
	channelsLock      *sync.RWMutex
	aliasesLock       *sync.RWMutex
}

func NewGraph() *Graph {
	return &Graph{
		Channels:          make(map[string]*Channel),
		Inbound:           make(map[string]map[string]Edge),
		Aliases:           make(map[string]string),
		adjacencyListLock: &sync.RWMutex{},
		channelsLock:      &sync.RWMutex{},
		aliasesLock:       &sync.RWMutex{},
	}
}

func (g *Graph) Lock() {
	g.channelsLock.Lock()
	g.adjacencyListLock.Lock()
	g.aliasesLock.Lock()
}

func (g *Graph) Unlock() {
	g.aliasesLock.Unlock()
	g.adjacencyListLock.Unlock()
	g.channelsLock.Unlock()
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
	g.channelsLock.Lock()
	g.adjacencyListLock.Lock()
	defer g.channelsLock.Unlock()
	defer g.adjacencyListLock.Unlock()

	// we need to do NewChannel and not only update the liquidity because of gossip updates
	for _, c := range channelList {
		var channel *Channel
		channelId := c.ShortChannelId + "/" + util.GetDirection(c.Source, c.Destination)
		// if the channel did not exist prior to this refresh estimate its initial liquidity to be 50/50
		if _, ok := g.Channels[channelId]; !ok {
			channel = NewChannel(c, uint64(0.5*float64(c.Satoshis*1000)), 0)
			g.AddChannel(channel)
		} else {
			channel = NewChannel(c, g.Channels[channelId].Liquidity, g.Channels[channelId].Timestamp)
		}
		g.Channels[channelId] = channel
	}
}

func (g *Graph) RefreshAliases(nodes []*glightning.Node) {
	g.aliasesLock.Lock()
	defer g.aliasesLock.Unlock()

	for _, n := range nodes {
		g.Aliases[n.Id] = n.Alias
	}
}

func (g *Graph) PruneChannels() {
	g.channelsLock.Lock()
	defer g.channelsLock.Unlock()

	// get current time in seconds
	now := uint(time.Now().Unix())

	// prune channels that are older than PRUNING_INTERVAL
	// TODO: remove closed channels, but might be worth waiting for glightning to implement channel_state_changed
	for _, c := range g.Channels {
		if c.LastUpdate+PRUNING_INTERVAL < now {
			g.DeleteChannel(c)
		}
	}
}

func (g *Graph) DeleteChannel(c *Channel) {
	// delete from channel map
	delete(g.Channels, c.ShortChannelId+"/"+util.GetDirection(c.Source, c.Destination))

	// delete from adjacency list
	for i, edge := range g.Inbound[c.Destination][c.Source] {
		if edge == c.ShortChannelId {
			g.Inbound[c.Destination][c.Source] = remove(g.Inbound[c.Destination][c.Source], i)
			break
		}
	}
}

// assumes valid input
func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (g *Graph) GetAlias(id string) string {
	g.aliasesLock.RLock()
	defer g.aliasesLock.RUnlock()

	if alias, ok := g.Aliases[id]; ok {
		return alias
	}
	return id
}

func (g *Graph) UpdateChannel(channelId, oppositeChannelId string, amount uint64) {
	g.channelsLock.Lock()
	defer g.channelsLock.Unlock()

	now := time.Now().Unix()

	if _, ok := g.Channels[channelId]; ok {
		g.Channels[channelId].Liquidity = amount
		g.Channels[channelId].Timestamp = now
	}

	if _, ok := g.Channels[oppositeChannelId]; ok {
		g.Channels[oppositeChannelId].Liquidity =
			g.Channels[oppositeChannelId].Satoshis*1000 - amount
		g.Channels[oppositeChannelId].Timestamp = now
	}
}

func (g *Graph) GetChannel(id string) (*Channel, error) {
	g.channelsLock.RLock()
	defer g.channelsLock.RUnlock()

	if _, ok := g.Channels[id]; !ok {
		return nil, util.ErrNoChannel
	}
	return g.Channels[id], nil
}

func (g *Graph) RefreshLiquidity(refreshThreshold time.Duration) int {
	g.channelsLock.Lock()
	defer g.channelsLock.Unlock()

	// get current time in seconds
	now := time.Now().Unix()
	hits := 0

	for _, c := range g.Channels {
		if c.Timestamp+int64(refreshThreshold.Seconds()) < now {
			c.ResetLiquidity()
			hits++
		}
	}

	return hits
}
