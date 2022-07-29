package parallel

import (
	"circular/rebalance"
	"circular/util"
)

const (
	DEFAULT_AMOUNT                = 400000
	DEFAULT_SPLITS                = 4
	DEFAULT_SPLIT_AMOUNT          = 100000
	DEFAULT_MAX_OUT_PPM           = 50
	DEFAULT_DEPLETE_UP_TO_PERCENT = 0.2
	DEFAULT_DEPLETE_UP_TO_AMOUNT  = 1000000
)

func (r *RebalanceParallel) setDefaults() {
	if r.Amount == 0 {
		r.Amount = DEFAULT_AMOUNT
	}
	if r.Splits <= 0 {
		r.Splits = DEFAULT_SPLITS
	}
	if r.SplitAmount == 0 {
		r.SplitAmount = DEFAULT_SPLIT_AMOUNT
	}
	if r.DepleteUpToPercent <= 0 {
		r.DepleteUpToPercent = DEFAULT_DEPLETE_UP_TO_PERCENT
	}
	if r.DepleteUpToAmount == 0 {
		r.DepleteUpToAmount = DEFAULT_DEPLETE_UP_TO_AMOUNT
	}
	if r.MaxOutPPM == 0 {
		r.MaxOutPPM = DEFAULT_MAX_OUT_PPM
	}
	if r.MaxPPM == 0 {
		r.MaxPPM = rebalance.DEFAULT_MAXPPM
	}
	if r.Attempts <= 0 {
		r.Attempts = rebalance.DEFAULT_ATTEMPTS
	}
	if r.MaxHops <= 0 {
		r.MaxHops = rebalance.DEFAULT_MAXHOPS
	}

	r.AmountRebalanced = 0
	r.InFlightAmount = 0

	// convert to msat
	r.DepleteUpToAmount *= 1000
	r.Amount *= 1000
	r.SplitAmount *= 1000
}

func (r *RebalanceParallel) validateParameters() error {
	if r.Amount < r.SplitAmount {
		return util.ErrAmountLessThanSplitAmount
	}
	if r.Amount%r.SplitAmount != 0 {
		return util.ErrAmountNotMultipleOfSplitAmount
	}
	if r.DepleteUpToPercent > 1 || r.DepleteUpToPercent < 0 {
		return util.ErrDepleteUpToPercentInvalid
	}
	return nil
}
