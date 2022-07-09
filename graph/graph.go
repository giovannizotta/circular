package graph

import (
	"circular/util"
	"container/heap"
	"encoding/json"
	"errors"
	"github.com/elementsproject/glightning/glightning"
	"log"
	"os"
	"time"
)

const (
	GRAPH_REFRESH = "10m"
	FILE          = "graph.json"
)

// Edge contains all the SCIDs of the channels going from nodeA to nodeB
type Edge []string

// Graph is the lightning network graph from the perspective of self
// It has been built from the gossip received by lightningd.
// To access the edges flowing out from a node, use: g.Outbound[node]
// To access an edge between nodeA and nodeB, use: g.Outbound[nodeA][nodeB]
// * an edge consists of one or more SCIDs between nodeA and nodeB
// To access a channel via channelId (scid/direction). use: g.Channels[channelId]
type Graph struct {
	Channels map[string]*Channel        `json:"channels"`
	Outbound map[string]map[string]Edge `json:"-"`
	Inbound  map[string]map[string]Edge `json:"-"`
}

func NewGraph() *Graph {
	var g *Graph
	g, err := loadFromFile()
	if err != nil {
		g = &Graph{
			Channels: make(map[string]*Channel),
		}
	}
	return g
}

func allocate(links *map[string]map[string]Edge, from, to string) {
	if *links == nil {
		*links = make(map[string]map[string]Edge)
	}
	if (*links)[from] == nil {
		(*links)[from] = make(map[string]Edge)
	}
	if (*links)[from][to] == nil {
		(*links)[from][to] = make([]string, 0)
	}
}

func (g *Graph) AddChannel(c *Channel) {
	allocate(&g.Outbound, c.Source, c.Destination)
	allocate(&g.Inbound, c.Destination, c.Source)
	g.Outbound[c.Source][c.Destination] = append(g.Outbound[c.Source][c.Destination], c.ShortChannelId)
	g.Inbound[c.Destination][c.Source] = append(g.Inbound[c.Destination][c.Source], c.ShortChannelId)
}

func (g *Graph) GetRoute(src, dst string, amount uint64, exclude map[string]bool) (*Route, error) {
	hops, err := g.dijkstra(src, dst, amount, exclude)
	if err != nil {
		return nil, err
	}
	route := NewRoute(src, dst, amount, hops)
	return route, nil
}

func (g *Graph) dijkstra(src, dst string, amount uint64, exclude map[string]bool) ([]RouteHop, error) {
	// start from the destination and find the source so that we can compute fees
	// TODO: consider that 32bits fees can be a problem but the api does it in that way
	log.Println("looking for a route from", src, "to", dst)
	distance := make(map[string]int)
	hop := make(map[string]RouteHop)
	maxDistance := 1 << 31
	for u := range g.Inbound {
		distance[u] = maxDistance
	}
	distance[dst] = 0
	log.Printf("distance map: %+v", distance)

	pq := make(PriorityQueue, 1, 16)
	// Insert destination
	pq[0] = &Item{value: &PqItem{
		Node:   dst,
		Amount: amount,
		Delay:  0,
	}, priority: 0}
	heap.Init(&pq)

	for pq.Len() > 0 {
		pqItem := heap.Pop(&pq).(*Item)
		u := pqItem.value.Node
		amount := pqItem.value.Amount
		delay := pqItem.value.Delay
		fee := pqItem.priority
		log.Printf("processing node %s with amount %d and delay %d", u, amount, delay)
		if u == src {
			break
		}
		if fee > distance[u] {
			continue
		}
		for v, edge := range g.Inbound[u] {
			if exclude[v] {
				continue
			}
			log.Println("checking edge", v)
			for _, scid := range edge {
				channel := g.Channels[scid+"/"+GetDirection(v, u)]
				log.Println("channel:", channel)
				if !channel.canUse(amount) {
					continue
				}
				log.Println("channel can be used")

				channelFee := int(channel.computeFee(amount))
				newDistance := distance[u] + channelFee
				if newDistance < distance[v] {
					log.Println("found new best fee coming from ", v, "with fee", newDistance)
					distance[v] = newDistance
					hop[v] = RouteHop{
						channel,
						amount,
						delay,
					}
					heap.Push(&pq, &Item{value: &PqItem{
						Node:   v,
						Amount: amount + uint64(channelFee),
						Delay:  delay + channel.Delay,
					}, priority: newDistance})
				}
			}
		}
	}
	if distance[src] == maxDistance {
		log.Println("no route found")
		return nil, errors.New("no route found")
	}
	// now we have the hop map, we can build the hops
	hops := make([]RouteHop, 0, 10)
	for u := src; u != dst; u = hop[u].Channel.Destination {
		hops = append(hops, hop[u])
	}
	return hops, nil
}

func loadFromFile() (*Graph, error) {
	defer util.TimeTrack(time.Now(), "graph.loadFromFile")
	file, err := os.Open(FILE)
	if err != nil {
		log.Println("unable to load from file:", err)
		log.Println("trying to load an old version of the graph")
		file, err = os.Open(FILE + ".old")
		if err != nil {
			log.Println("unable to load any old version of the file: ", err)
			return nil, err
		}
	}
	defer file.Close()
	g := &Graph{
		Channels: make(map[string]*Channel),
	}
	err = json.NewDecoder(file).Decode(&g)
	if err != nil {
		return nil, err
	}
	// TODO: add Outbound and Inbound
	return g, nil
}

func (g *Graph) SaveToFile() {
	defer util.TimeTrack(time.Now(), "graph.SaveToFile")
	// open temporary file
	file, err := os.Create(FILE + ".tmp")
	if err != nil {
		log.Printf("error opening file: %v", err)
		return
	}
	defer file.Close()
	// write json
	err = json.NewEncoder(file).Encode(g)
	if err != nil {
		log.Printf("error writing file: %v", err)
		return
	}

	// save old file
	// check if FILE exists
	if _, err := os.Stat(FILE); err == nil {
		err = os.Rename(FILE, FILE+".old")
	}
	// rename tmp to FILE
	err = os.Rename(FILE+".tmp", FILE)
}

func (g *Graph) Refresh(channelList []*glightning.Channel) {
	defer util.TimeTrack(time.Now(), "graph.Refresh")
	for _, c := range channelList {
		var channel *Channel
		channelId := c.ShortChannelId + "/" + GetDirection(c.Source, c.Destination)
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

func GetDirection(from, to string) string {
	if from < to {
		return "1"
	}
	return "0"
}
