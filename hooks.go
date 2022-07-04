package main

import (
	"github.com/elementsproject/glightning/glightning"
	"log"
)

func registerHooks(p *glightning.Plugin) {
	p.RegisterHooks(&glightning.Hooks{
		HtlcAccepted: OnHtlcAccepted,
	})
}

func OnHtlcAccepted(event *glightning.HtlcAcceptedEvent) (*glightning.HtlcAcceptedResponse, error) {
	log.Printf("htlc: %+v\n", event.Htlc)

	if r, ok := ongoingRebalances[event.Htlc.PaymentHash]; ok {
		log.Println("found an htlc which we can resolve")
		return event.Resolve(r.PreimageHashPair.Preimage), nil
	}

	return event.Continue(), nil
}
