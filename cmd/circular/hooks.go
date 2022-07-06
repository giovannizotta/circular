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
	self := node.GetSelf()
	log.Printf("htlc: %+v\n", event.Htlc)

	if r, ok := self.OngoingRebalances[event.Htlc.PaymentHash]; ok {
		log.Println("found an htlc which we can resolve")
		return event.Resolve(r), nil
	}

	return event.Continue(), nil
}
