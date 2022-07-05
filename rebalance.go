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
	MaxPPM           uint64           `json:"max_ppm"`
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

func getBestPeerChannel(peer string, metric func(channel *glightning.PeerChannel) uint64) *glightning.PeerChannel {
	channels := getPeerChannels(peer)
	best := channels[0]
	for _, channel := range channels {
		if metric(channel) > metric(best) {
			best = channel
		}
	}
	return best
}

func (r *Rebalance) validatePeerParameters() error {
	//validate that the nodes are not self
	if r.In == self.Id || r.Out == self.Id {
		return errors.New("one of the nodes is self")
	}
	//validate that the nodes are not the same
	if r.In == r.Out {
		return errors.New("incoming and outgoing nodes are the same")
	}
	//validate that the r.Destination is a neighbor of self
	if _, ok := graph.Outbound[self.Id][r.In]; !ok {
		return errors.New("incoming value is not a peer")
	}
	//validate r.Source is in graph.Outbound[self]
	if _, ok := graph.Outbound[self.Id][r.Out]; !ok {
		return errors.New("outgoing value is not a peer")
	}
	if len(self.Peers) == 0 {
		return errors.New("no peers yet")
	}
	return nil
}

func (r *Rebalance) validateLiquidityParameters() error {
	inChannel := getBestPeerChannel(r.In, func(channel *glightning.PeerChannel) uint64 {
		return channel.ReceivableMilliSatoshi
	})
	outChannel := getBestPeerChannel(r.Out, func(channel *glightning.PeerChannel) uint64 {
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

func (r *Rebalance) validateParameters() error {
	if r.In == "" || r.Out == "" || r.Amount <= 0 || r.MaxPPM < 0 {
		return errors.New("missing required parameter")
	}
	err := r.validatePeerParameters()
	if err != nil {
		return err
	}

	err = r.validateLiquidityParameters()
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

func sendPay(route []glightning.RouteHop, paymentHash string) (*glightning.SendPayFields, error) {
	_, err := lightning.SendPayLite(route, paymentHash)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//TODO: learn from failed payments
	result, err := lightning.WaitSendPay(paymentHash, 20)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return result, nil
}

func (r *Rebalance) getRoute() (*Route, error) {
	//exclude := r.excludeEdgesToSelf()
	route, err := NewRoute(r.In, r.Out, r.Amount, []string{self.Id})
	if err != nil {
		return nil, err
	}

	// prepend self.Id to the beginning and to the end
	route.prependHop(self.Id)
	route.appendHop(self.Id)
	for i, hop := range route.Hops {
		log.Printf("hop %d: %+v\n", i, hop)
	}

	if route.FeePPM > r.MaxPPM {
		return nil, errors.New(fmt.Sprintf("route too expensive. "+
			"Cheapest route found was %d ppm, but max_ppm is %d",
			route.FeePPM/1000, r.MaxPPM/1000))
	}
	return route, nil
}

func (r *Rebalance) run() (string, error) {
	//TODO: save the preimage hash pair in the database
	log.Println("generating preimage/hash pair")
	r.PreimageHashPair = *NewPreimageHashPair()
	ongoingRebalances[r.PreimageHashPair.Hash] = *r

	log.Println("searching for a route")
	route, err := r.getRoute()
	if err != nil {
		return "", err
	}

	log.Println("trying to send payment to route")
	_, err = sendPay(route.Hops, r.PreimageHashPair.Hash)
	if err != nil {
		return "", err
	}
	r.clean()

	//TODO: after successful rebalance, refresh channel balances
	return fmt.Sprintf("rebalance successful at %d ppm\n", route.FeePPM/1000), nil
}

func (r *Rebalance) Call() (jrpc2.Result, error) {
	err := r.validateParameters()
	if err != nil {
		return nil, err
	}
	log.Printf("parameters validated, running rebalance\n")
	//convert to msatoshi
	r.Amount *= 1000
	r.MaxPPM *= 1000

	result, err := r.run()
	if err != nil {
		return nil, err
	}
	return NewRebalanceResult(result), nil
}

func (r *Rebalance) clean() {
	delete(ongoingRebalances, r.PreimageHashPair.Hash)
}
