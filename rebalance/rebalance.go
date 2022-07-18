package rebalance

import (
	"circular/graph"
	"circular/node"
	"circular/util"
	"errors"
	"fmt"
	"log"
	"strconv"
)

const (
	NORMAL           = "CHANNELD_NORMAL"
	DEFAULT_AMOUNT   = 200000000
	DEFAULT_PPM      = 100
	DEFAULT_ATTEMPTS = 1
	MAX_HOPS         = 15
)

type Rebalance struct {
	OutChannel *graph.Channel
	InChannel  *graph.Channel
	Amount     uint64
	MaxPPM     uint64
	Attempts   int
	Node       *node.Node
}

func NewRebalance(outChannel, inChannel *graph.Channel, amount, maxppm uint64, attempts int) *Rebalance {
	return &Rebalance{
		OutChannel: outChannel,
		InChannel:  inChannel,
		Amount:     amount,
		MaxPPM:     maxppm,
		Attempts:   attempts,
		Node:       node.GetNode(),
	}
}

func (r *Rebalance) Setup() error {
	err := r.setDefaultParameters()
	if err != nil {
		return err
	}

	err = r.validateLiquidityParameters(r.OutChannel, r.InChannel)
	if err != nil {
		return err
	}
	return nil
}

func (r *Rebalance) Run() (*Result, error) {
	maxHops := 3
	var err error
	var result string
	var i int
	for i = 1; i <= r.Attempts; {
		if maxHops > MAX_HOPS {
			err = errors.New("unable to find a route with less than " +
				strconv.Itoa(MAX_HOPS) +
				" hops: " + err.Error())
			break
		}
		log.Println("===================== ATTEMPT", i, "=====================")
		result, err = r.runAttempt(maxHops, r.OutChannel, r.InChannel)

		// success
		if err == nil {
			result += " after " + strconv.Itoa(i) + " attempts"
			log.Println(result)
			return NewResult(result), nil
		}

		// no route found with at most maxHops
		if err == util.ErrNoRoute {
			log.Println("no route found with at most", maxHops, "hops, increasing max hops to ", maxHops+1)
			maxHops += 1
			continue
		}

		// no route found with at most maxHops cheaper than maxPPM
		if errors.As(err, &util.ErrRouteTooExpensive{}) {
			log.Println(err, ", increasing max hops to ", maxHops+1)
			maxHops += 1
			continue
		}

		// sendpay timeout
		if err == util.ErrSendPayTimeout {
			err = errors.New("rebalancing timed out after " +
				strconv.Itoa(node.SENDPAY_TIMEOUT) +
				"s. The payment is still in flight and may still succeed.")
			break
		}

		// TODO: handle case where the peer channel has gone offline
		if err != util.ErrTemporaryFailure {
			break
		}
		i++
	}
	return NewResult("rebalance failed after " + strconv.Itoa(i) + " attempts, last error: " + err.Error()), nil
}

func (r *Rebalance) runAttempt(maxHops int, outgoingChannel *graph.Channel, incomingChannel *graph.Channel) (string, error) {
	route, err := r.tryRoute(maxHops, outgoingChannel, incomingChannel)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(""+
			"successfully rebalanced %d sats "+
			"from %s to %s at %d ppm. Total fees paid: %.3f sats",
			r.Amount/1000,
			r.Node.Graph.Aliases[outgoingChannel.Destination], r.Node.Graph.Aliases[incomingChannel.Source],
			route.FeePPM(), float64(route.Fee())/1000),
		nil
}
