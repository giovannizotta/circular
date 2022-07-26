package main

import (
	"circular/node"
	"github.com/elementsproject/glightning/glightning"
)

func registerHooks(p *glightning.Plugin) {
	p.RegisterHooks(&glightning.Hooks{
		HtlcAccepted: OnHtlcAccepted,
	})
}

func OnHtlcAccepted(event *glightning.HtlcAcceptedEvent) (*glightning.HtlcAcceptedResponse, error) {
	self := node.GetNode()
	preimage, err := self.DB.Get(event.Htlc.PaymentHash)
	if err != nil {
		return event.Continue(), nil
	}
	self.Logln(glightning.Info, "resolving HTLC with preimage:", string(preimage))
	return event.Resolve(string(preimage)), nil
}
