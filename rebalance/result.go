package rebalance

import "github.com/elementsproject/glightning/glightning"

type Result struct {
	Result     string `json:"result"`
	FormatHint string `json:"format-hint,omitempty"`
}

func NewResult(result string) *Result {
	return &Result{
		Result:     result,
		FormatHint: glightning.FormatSimple,
	}
}
