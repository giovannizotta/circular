package parallel

type Result struct {
	Status           string `json:"status"`
	Message          string `json:"message"`
	RebalanceTarget  uint64 `json:"rebalance_target"`
	RebalancedAmount uint64 `json:"rebalanced_amount"`
	Attempts         int    `json:"attempts"`
}

func NewResult(rebalance *RebalanceParallel) *Result {
	return &Result{
		Status:           "success",
		Message:          "rebalance completed",
		RebalanceTarget:  rebalance.Amount / 1000,
		RebalancedAmount: rebalance.AmountRebalanced / 1000,
		Attempts:         rebalance.TotalAttempts,
	}
}
