package rebalance

import (
	"errors"
	"fmt"
)

var (
	ErrTemporaryFailure = errors.New("TEMPORARY_FAILURE")
	ErrSendPayTimeout   = errors.New("200:Timed out while waiting")
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
