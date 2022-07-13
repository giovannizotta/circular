package main

import (
	"circular/node"
	"github.com/elementsproject/glightning/glightning"
	"log"
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
	log.Println("resolving HTLC with preimage:", preimage)
	return event.Resolve(preimage), nil
}
