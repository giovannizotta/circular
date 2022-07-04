package main

import (
	"errors"
	"fmt"
	"github.com/elementsproject/glightning/glightning"
	"github.com/elementsproject/glightning/jrpc2"
	"log"
)

const (
	NORMAL = "CHANNELD_NORMAL"
)

type Rebalance struct {
	In               string           `json:"in"`
	Out              string           `json:"out"`
	Amount           uint64           `json:"amount"`
	PreimageHashPair PreimageHashPair `json:"preimage,omitempty"`
}

func (r *Rebalance) Name() string {
	return "rebalance"
}

func (r *Rebalance) New() interface{} {
	return &Rebalance{}
}

func getPeerChannels(id string) []*glightning.PeerChannel {
	peer, err := lightning.GetPeer(id)
	if err != nil {
		log.Fatalln(err)
	}
	return peer.Channels
}

func getBestChannel(peer string, metric func(channel *glightning.PeerChannel) uint64) *glightning.PeerChannel {
	channels := getPeerChannels(peer)
	best := channels[0]
	for _, channel := range channels {
		if metric(channel) > metric(best) {
			best = channel
		}
	}
	return best
}

func (r *Rebalance) validateRebalancePeerParameters() error {
	//validate that the nodes are not self
	if r.In == self.Id || r.Out == self.Id {
		return errors.New("one of the nodes is self")
	}
	//validate that the nodes are not the same
	if r.In == r.Out {
		return errors.New("incoming and outgoing nodes are the same")
	}
	//validate that the r.In is a neighbor of self
	if _, ok := graph.Nodes[self.Id][r.In]; !ok {
		return errors.New("incoming node is not a peer")
	}
	//validate r.Out is in graph.Nodes[self]
	if _, ok := graph.Nodes[self.Id][r.Out]; !ok {
		return errors.New("outgoing node is not a peer")
	}
	if len(self.Peers) == 0 {
		return errors.New("no peers yet")
	}
	return nil
}

func (r *Rebalance) validateRebalanceLiquidityParameters() error {
	inChannel := getBestChannel(r.In, func(channel *glightning.PeerChannel) uint64 {
		return channel.ReceivableMilliSatoshi
	})
	outChannel := getBestChannel(r.Out, func(channel *glightning.PeerChannel) uint64 {
		return channel.SpendableMilliSatoshi
	})
	//validate that the channels are in normal state
	if inChannel.State != NORMAL {
		return errors.New("incoming channel is not in normal state")
	}
	if outChannel.State != NORMAL {
		return errors.New("outgoing channel is not in normal state")
	}
	//validate that the amount is less than the liquidity of the channels
	if (inChannel.ReceivableMilliSatoshi) < r.Amount {
		return errors.New("incoming channel has insufficient remote balance")
	}
	if (outChannel.SpendableMilliSatoshi) < r.Amount {
		return errors.New("outgoing channel has insufficient local balance")
	}
	return nil
}

func (r *Rebalance) validateRebalanceParameters() error {
	if r.In == "" || r.Out == "" || r.Amount <= 0 {
		return errors.New("missing required parameter")
	}
	err := r.validateRebalancePeerParameters()
	if err != nil {
		return err
	}

	err = r.validateRebalanceLiquidityParameters()
	if err != nil {
		return err
	}

	return nil
}

type RebalanceResult struct {
	Result     string `json:"rebalance"`
	FormatHint string `json:"format-hint,omitempty"`
}

func NewRebalanceResult(result string) *RebalanceResult {
	return &RebalanceResult{
		Result:     result,
		FormatHint: glightning.FormatSimple,
	}
}

func getDirection(from string, to string) uint8 {
	if from < to {
		return 0
	}
	return 1
}

func (r *Rebalance) computeFirstHopFeeMillisatoshi() uint64 {
	baseFee := graph.Nodes[self.Id][r.Out][0].BaseFeeMillisatoshi
	result := baseFee
	proportionalFee := ((r.Amount / 1000) * graph.Nodes[self.Id][r.Out][0].FeePerMillionth) / 1000
	result += proportionalFee
	log.Printf("base fee: %d, proportional fee: %d, result: %d\n", baseFee, proportionalFee, result)
	return result
}

func (r *Rebalance) prependHop(route []glightning.RouteHop) []glightning.RouteHop {
	//FIXME: get the best channel?
	routeHop := glightning.RouteHop{
		Id:             r.Out,
		ShortChannelId: graph.Nodes[self.Id][r.Out][0].ShortChannelId,
		MilliSatoshi:   route[0].MilliSatoshi + r.computeFirstHopFeeMillisatoshi(),
		Delay:          route[0].Delay + graph.Nodes[self.Id][r.Out][0].Delay,
		Direction:      getDirection(self.Id, r.Out),
	}
	//prepend the hop to the route
	route = append([]glightning.RouteHop{routeHop}, route...)
	return route
}

func (r *Rebalance) getExcludeList() []string {
	var result []string
	var alreadyExcluded map[string]bool
	alreadyExcluded = make(map[string]bool)
	for _, channel := range graph.Nodes[self.Id][r.Out] {
		//exclude direction out -> self
		candidate := channel.ShortChannelId + "/"
		if self.Id < r.Out {
			candidate += "1"
		} else {
			candidate += "0"
		}
		if _, found := alreadyExcluded[candidate]; !found {
			result = append(result, candidate)
		}
		alreadyExcluded[candidate] = true
	}
	return result
}

func (r *Rebalance) buildRoute() (*[]glightning.RouteHop, error) {
	exclude := r.getExcludeList()

	route, err := lightning.GetRoute(self.Id, r.Amount, 20, 0, r.Out, 5, exclude, 20)
	if err != nil {
		return nil, err
	}

	for i := range route {
		route[i].AmountMsat = ""
	}

	route = r.prependHop(route)
	return &route, nil
}

func (r *Rebalance) sendPayToRoute(route *[]glightning.RouteHop) (*glightning.SendPayFields, error) {
	_, err := lightning.SendPayLite(*route, r.PreimageHashPair.Hash)
	if err != nil {
		log.Println(err)
	}

	result, err := lightning.WaitSendPay(r.PreimageHashPair.Hash, 20)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Rebalance) run() (string, error) {
	log.Println("building route")
	route, err := r.buildRoute()
	if err != nil {
		return "", err
	}
	log.Println("sending payment to route")
	result, err := r.sendPayToRoute(route)
	if err != nil {
		return "", err
	}

	log.Printf("payment successful: %+v\n", result)
	r.finish()

	return fmt.Sprintf("rebalance successful\n"), nil
}

func (r *Rebalance) finish() {
	delete(ongoingRebalances, r.PreimageHashPair.Hash)
}

func (r *Rebalance) Call() (jrpc2.Result, error) {
	err := r.validateRebalanceParameters()
	if err != nil {
		return nil, err
	}
	log.Printf("parameters validated, running rebalance\n")
	r.Amount *= 1000

	r.PreimageHashPair = *NewPreimageHashPair()
	ongoingRebalances[r.PreimageHashPair.Hash] = r

	result, err := r.run()
	if err != nil {
		return nil, err
	}
	return NewRebalanceResult(result), nil
}
