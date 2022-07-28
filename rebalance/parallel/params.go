package parallel

import (
	"circular/rebalance"
	"circular/util"
)

const (
	SPLITS                = 2
	SPLIT_AMOUNT          = 100000
	MAX_OUT_PPM           = 10
	DEPLETE_UP_TO_PERCENT = 0.5
	DEPLETE_UP_TO_AMOUNT  = 2000000
)

func (r *RebalanceParallel) setDefaults() {
	if r.Splits <= 0 {
		r.Splits = SPLITS
	}
	if r.SplitAmount == 0 {
		r.SplitAmount = SPLIT_AMOUNT
	}
	if r.Amount == 0 {
		r.Amount = SPLITS * SPLIT_AMOUNT
	}
	if r.DepleteUpToPercent <= 0 {
		r.DepleteUpToPercent = DEPLETE_UP_TO_PERCENT
	}
	if r.DepleteUpToAmount == 0 {
		r.DepleteUpToAmount = DEPLETE_UP_TO_AMOUNT
	}
	if r.MaxOutPPM == 0 {
		r.MaxOutPPM = MAX_OUT_PPM
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
