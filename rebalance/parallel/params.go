package parallel

import (
	"circular/rebalance"
	"circular/util"
)

const (
	DEFAULT_AMOUNT       = 400000
	DEFAULT_SPLITS       = 4
	DEFAULT_SPLIT_AMOUNT = 100000
)

func (r *AbstractRebalance) setGenericDefaults() {
	if r.amount == 0 {
		r.amount = DEFAULT_AMOUNT
	}
	if r.splits == 0 {
		r.splits = DEFAULT_SPLITS
	}
	if r.splitAmount == 0 {
		r.splitAmount = DEFAULT_SPLIT_AMOUNT
	}
	if r.maxPPM == 0 {
		r.maxPPM = rebalance.DEFAULT_MAXPPM
	}
	if r.attempts <= 0 {
		r.attempts = rebalance.DEFAULT_ATTEMPTS
	}
	if r.maxHops <= 0 {
		r.maxHops = rebalance.DEFAULT_MAXHOPS
	}

	r.AmountRebalanced = 0
	r.InFlightAmount = 0

	// convert to msat
	r.amount *= 1000
	r.splitAmount *= 1000
}

func (r *AbstractRebalance) validateGenericParameters() error {
	if r.amount < r.splitAmount {
		return util.ErrAmountLessThanSplitAmount
	}
	if r.amount%r.splitAmount != 0 {
		return util.ErrAmountNotMultipleOfSplitAmount
	}
	return nil
}
