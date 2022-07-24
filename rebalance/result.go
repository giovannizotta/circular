package rebalance

import "circular/graph"

type Result struct {
	Status     string             `json:"status"`
	Message    string             `json:"message"`
	Amount     uint64             `json:"amount"`
	Out        string             `json:"out"`
	In         string             `json:"in"`
	Attempts   uint64             `json:"attempts"`
	Fee        uint64             `json:"fee,omitempty"`
	PPM        uint64             `json:"ppm,omitempty"`
	Route      *graph.PrettyRoute `json:"route,omitempty"`
	FormatHint string             `json:"format-hint,omitempty"`
}

func NewResult(status string, amount uint64, src, dst string) *Result {
	return &Result{
		Status: status,
		Amount: amount,
		Out:    src,
		In:     dst,
	}
}
