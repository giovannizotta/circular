package rebalance

import (
	"github.com/elementsproject/glightning/glightning"
)

type Result struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	Amount     uint64 `json:"amount"`
	Out        string `json:"out"`
	In         string `json:"in"`
	Fee        uint64 `json:"fee"`
	PPM        uint64 `json:"ppm"`
	Attempts   uint64 `json:"attempts"`
	Hops       uint64 `json:"hops"`
	FormatHint string `json:"format-hint,omitempty"`
}

/*
r.Amount/1000,
			r.Node.Graph.Aliases[r.OutChannel.Destination], r.Node.Graph.Aliases[r.InChannel.Source],
			route.FeePPM(), float64(route.Fee())/1000
*/
func NewResult(status string, amount uint64, src, dst string, fee, ppm, hops uint64) *Result {
	return &Result{
		Status:     status,
		Amount:     amount,
		Out:        src,
		In:         dst,
		Fee:        fee,
		PPM:        ppm,
		Hops:       hops,
		FormatHint: glightning.FormatSimple,
	}
}
