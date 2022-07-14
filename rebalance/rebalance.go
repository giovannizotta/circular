package rebalance

import (
	"circular/graph"
	"circular/node"
	"circular/util"
	"errors"
	"fmt"
	"github.com/elementsproject/glightning/jrpc2"
	"log"
	"strconv"
	"time"
)

const (
	NORMAL           = "CHANNELD_NORMAL"
	DEFAULT_AMOUNT   = 200000000
	DEFAULT_PPM      = 100
	DEFAULT_ATTEMPTS = 10
)

type Rebalance struct {
	Out      string     `json:"out"`
	In       string     `json:"in"`
	Amount   uint64     `json:"amount,omitempty"`
	MaxPPM   uint64     `json:"maxppm,omitempty"`
	Attempts int        `json:"attempts,omitempty"`
	Node     *node.Node `json:"-"`
}

func (r *Rebalance) Name() string {
	return "circular"
}

func (r *Rebalance) New() interface{} {
	return &Rebalance{}
}

func (r *Rebalance) Call() (jrpc2.Result, error) {
	r.Node = node.GetNode()
	//convert to msatoshi
	r.Amount *= 1000

	log.Println("circular rebalance called")
	log.Println("self: ", r.Node.Id)
	log.Println("in:", r.In)
	log.Println("out:", r.Out)
	log.Println("amount:", r.Amount, "maxppm:", r.MaxPPM)
	log.Println("attempts:", r.Attempts)
	if err := r.validateParameters(); err != nil {
		return nil, err
	}

	maxHops := 3
	var err error
	var result string
	for i := 1; i <= r.Attempts; i++ {
		log.Println("===================== ATTEMPT", i, "=====================")
		result, err = r.run(maxHops)
		if err == nil {
			result += " after " + strconv.Itoa(i) + " attempts"
			log.Println(result)
			return NewResult(result), nil
		}
		if err == graph.ErrNoRoute {
			log.Println("no route found with at most", maxHops, "hops, increasing max hops to ", maxHops+1)
			maxHops += 1
			continue
		}
		if errors.As(err, &RouteTooExpensiveError{}) {
			log.Println(err, "increasing max hops to ", maxHops+1)
			maxHops += 1
			continue
		}
		// TODO: handle case where the peer channel has gone offline
		if err != TemporaryFailureError {
			return NewResult(result), nil
		}
	}

	return NewResult("rebalance failed after " + strconv.Itoa(r.Attempts) + " attempts, last error: " + err.Error()), nil
}

func (r *Rebalance) run(maxHops int) (string, error) {
	defer util.TimeTrack(time.Now(), "rebalance.run")

	route, err := r.tryRoute(maxHops)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(""+
		"successfully rebalanced %d sats "+
		"from %s to %s at %d ppm. Total fees paid: %.3f sats",
		r.Amount/1000, r.Out[:8], r.In[:8], route.FeePPM(), float64(route.Fee())/1000), nil
}
