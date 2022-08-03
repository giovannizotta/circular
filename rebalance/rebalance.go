package rebalance

import (
	"circular/graph"
	"circular/node"
	"circular/util"
	"errors"
	"fmt"
	"github.com/elementsproject/glightning/glightning"
	"strconv"
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
	r.setDefaults()

	if err := r.validateLiquidityParameters(r.OutChannel, r.InChannel); err != nil {
		return err
	}

	return nil
}

func (r *Rebalance) Run() *Result {
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
		r.Node.Logln(glightning.Debug, "===================== ATTEMPT ", i, " =====================")

		result, err := r.runAttempt(maxHops)

		// success
		if err == nil {
			result.Attempts = uint64(i)
			r.Node.Logln(glightning.Debug, result)
			return result
		}

		// no route found with at most maxHops
		if err == util.ErrNoRoute {
			r.Node.Logln(glightning.Debug, "no route found with at most ", maxHops, " hops, increasing max hops to ", maxHops+1)
			lastError = err.Error()
			maxHops += 1
			continue
		}

		// no route found with at most maxHops cheaper than maxPPM
		if errors.As(err, &util.ErrRouteTooExpensive{}) {
			r.Node.Logln(glightning.Debug, err, ", increasing max hops to ", maxHops+1)
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

	return failure
}

func (r *Rebalance) runAttempt(maxHops int) (*Result, error) {
	if err := r.validateLiquidityParameters(r.OutChannel, r.InChannel); err != nil {
		return nil, err
	}

	route, err := r.tryRoute(maxHops)
	if err != nil {
		return nil, err
	}

	result := NewResult("success", r.Amount/1000,
		r.OutChannel.Destination, r.InChannel.Source)

	result.Fee = route.Fee
	result.PPM = route.FeePPM
	result.Route = route
	result.Message = fmt.Sprintf("successfully rebalanced %d sats from %s to %s at %d ppm. Total fees paid: %.3f sats",
		result.Amount, r.Node.Graph.GetAlias(r.OutChannel.Destination), r.Node.Graph.GetAlias(r.InChannel.Source),
		result.PPM, float64(result.Fee)/1000)

	return result, nil
}
