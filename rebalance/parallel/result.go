package parallel

import (
	"circular/rebalance"
)

// Success is a map from PPM to amount
// for each PPM, the amount of sats rebalanced at that ppm
type Success map[uint64]uint64

type Result struct {
	RebalanceTarget  uint64             `json:"rebalance_target"`
	RebalancedAmount uint64             `json:"rebalanced_amount"`
	Attempts         uint64             `json:"attempts"`
	Time             string             `json:"time"`
	Successes        map[string]Success `json:"successes"`
}

func NewResult(target uint64) *Result {
	return &Result{
		RebalanceTarget:  target / 1000,
		RebalancedAmount: 0,
		Successes:        make(map[string]Success),
	}
}

func (r *Result) AddSuccess(result *rebalance.Result, aliases map[string]string) {
	r.RebalancedAmount += result.Amount
	alias := aliases[result.Out]
	if _, ok := r.Successes[alias]; !ok {
		r.Successes[alias] = make(map[uint64]uint64)
	}
	if _, ok := r.Successes[alias][result.PPM]; !ok {
		r.Successes[alias][result.PPM] = 0
	}

	r.Successes[alias][result.PPM] += result.Amount
}
