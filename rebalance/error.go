package rebalance

import (
	"errors"
	"fmt"
)

var (
	TemporaryFailureError = errors.New("TEMPORARY_FAILURE")
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
