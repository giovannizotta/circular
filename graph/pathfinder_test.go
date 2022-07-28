package graph

import (
	"circular/util"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func LoadGraphFromFile(dir, filename string) (*Graph, error) {
	file, err := os.Open(dir + "/" + filename)
	if err != nil {
		if err != nil {
			return nil, util.ErrNoGraphToLoad
		}
	}
	defer file.Close()

	g := NewGraph()

	err = json.NewDecoder(file).Decode(g)
	if err != nil {
		return nil, err
	}

	for _, c := range g.Channels {
		g.AddChannel(c)
	}
	return g, nil
}

func TestPathfinderBasic(t *testing.T) {
	t.Log("graph/pathfinder_test.go")

	graph, err := LoadGraphFromFile("testdata", "graph.json")
	if err != nil {
		t.Fatal(err)
	}
	src := "02d41224b71a5346a656f8949c66d11495e39dac55ab8772f55c26ca515db910ea"
	dst := "03c731efa9935d869d87e57d4496de2b3badfb9ec7dbbd40051fb19351027336c5"
	amount := 200000000
	exclude := map[string]bool{
		"02a30b35b374b0bde273f2e36f1a6db9b1d9f4591d00416ffa541b6eb16e70921f": true,
	}
	maxHops := 10

	hops, err := graph.dijkstra(src, dst, uint64(amount), exclude, maxHops)
	if err != nil {
		t.Fatal(err)
	}
	assert.LessOrEqual(t, len(hops), maxHops)
	for i := 0; i < len(hops)-1; i++ {
		assert.Equal(t, hops[i].Destination, hops[i+1].Source)
		assert.GreaterOrEqual(t, hops[i].Liquidity, hops[i].MilliSatoshi)
		assert.GreaterOrEqual(t, hops[i].MilliSatoshi, hops[i+1].MilliSatoshi)
		assert.Greater(t, hops[i].Delay, hops[i+1].Delay)
	}
	assert.Equal(t, hops[len(hops)-1].Destination, dst)
	assert.Equal(t, hops[0].Source, src)
}
