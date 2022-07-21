package util

import (
	"errors"
	"fmt"
)

type ErrRouteTooExpensive struct {
	FeePPM uint64
	MaxPPM uint64
}

func NewRouteTooExpensiveError(feePPM uint64, maxPPM uint64) ErrRouteTooExpensive {
	return ErrRouteTooExpensive{
		FeePPM: feePPM,
		MaxPPM: maxPPM,
	}
}

func (e ErrRouteTooExpensive) Error() string {
	return fmt.Sprintf("route too expensive. Cheapest route found was %d ppm, but maxppm is %d", e.FeePPM, e.MaxPPM)
}

var (
	ErrSelfNode            = errors.New("one of the nodes is self")
	ErrNoPeers             = errors.New("no peers yet")
	ErrNoRequiredParameter = errors.New("missing required parameter")
	ErrNoPeerChannel       = errors.New("not a peer or peer channel")
	ErrTemporaryFailure    = errors.New("TEMPORARY_FAILURE")
	ErrSendPayTimeout      = errors.New("200:Timed out while waiting")
	ErrNoSuchNode          = errors.New("no such node")
	ErrNoRoute             = errors.New("no route")
	ErrNoOutgoingChannel   = errors.New("no outgoing channel")
	ErrNoIncomingChannel   = errors.New("no incoming channel")
	ErrNoGraphToLoad       = errors.New("no graph to load")
)
