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
	DEFAULT_MAXPPM   = 10
	DEFAULT_ATTEMPTS = 1
	DEFAULT_MAXHOPS  = 8
)

type Rebalance struct {
	OutChannel *graph.Channel
	InChannel  *graph.Channel
	Amount     uint64
	MaxPPM     uint64
	Attempts   int
	MaxHops    int
	Node       *node.Node
}

func NewRebalance(outChannel, inChannel *graph.Channel, amount, maxppm uint64, attempts, maxHops int) *Rebalance {
	return &Rebalance{
		OutChannel: outChannel,
		InChannel:  inChannel,
		Amount:     amount,
		MaxPPM:     maxppm,
		Attempts:   attempts,
		MaxHops:    maxHops,
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
	var (
		maxHops   = 3
		i         = 1
		lastError = ""
	)
	for i <= r.Attempts {
		if maxHops > r.MaxHops {
			lastError = " Unable to find a route with less than " +
				strconv.Itoa(r.MaxHops) + " hops. " + lastError
			break
		}
		log.Println("===================== ATTEMPT", i, "=====================")

		result, err := r.runAttempt(maxHops)

		// success
		if err == nil {
			result.Attempts = uint64(i)
			log.Println(result)
			return result, nil
		}

		// no route found with at most maxHops
		if err == util.ErrNoRoute {
			log.Println("no route found with at most", maxHops, "hops, increasing max hops to ", maxHops+1)
			lastError = err.Error()
			maxHops += 1
			continue
		}

		// no route found with at most maxHops cheaper than maxPPM
		if errors.As(err, &util.ErrRouteTooExpensive{}) {
			log.Println(err, ", increasing max hops to ", maxHops+1)
			lastError = err.Error()
			maxHops += 1
			continue
		}

		// sendpay timeout
		if err == util.ErrSendPayTimeout {
			lastError = "rebalancing timed out after " +
				strconv.Itoa(node.SENDPAY_TIMEOUT) +
				"s."
			break
		}

		// TODO: handle case where the peer channel has gone offline (First peer not ready)
		if err != util.ErrTemporaryFailure {
			lastError = err.Error()
			break
		}
		i++
	}

	failure := NewResult("failure", r.Amount/1000, r.OutChannel.Destination, r.InChannel.Source)
	failure.Attempts = uint64(i - 1)
	failure.Message = "rebalance failed after " + strconv.Itoa(int(failure.Attempts)) + " attempts."
	failure.Message += lastError

	return failure, nil
}

func (r *Rebalance) runAttempt(maxHops int) (*Result, error) {
	// refresh the peer channels in case they have changed fees/liquidity/state/whatever
	r.Node.RefreshPeer(r.OutChannel.Destination)
	r.Node.RefreshPeer(r.InChannel.Source)

	err := r.validateLiquidityParameters(r.OutChannel, r.InChannel)
	if err != nil {
		return nil, err
	}

	route, err := r.tryRoute(maxHops)
	if err != nil {
		return nil, err
	}

	result := NewResult("success", r.Amount/1000,
		r.OutChannel.Destination, r.InChannel.Source)

	// get aliases, if any
	var srcAlias, dstAlias string
	if _, ok := r.Node.Graph.Aliases[r.OutChannel.Destination]; ok {
		srcAlias = r.Node.Graph.Aliases[r.OutChannel.Destination]
	} else {
		srcAlias = r.OutChannel.Destination
	}
	if _, ok := r.Node.Graph.Aliases[r.InChannel.Source]; ok {
		dstAlias = r.Node.Graph.Aliases[r.InChannel.Source]
	} else {
		dstAlias = r.InChannel.Source
	}

	result.Fee = route.Fee()
	result.PPM = route.FeePPM()
	result.Route = graph.NewPrettyRoute(route)
	result.Message = fmt.Sprintf("successfully rebalanced %d sats from %s to %s at %d ppm. Total fees paid: %.3f sats",
		result.Amount, srcAlias, dstAlias,
		result.PPM, float64(result.Fee)/1000)

	return result, nil
}
