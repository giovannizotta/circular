package main

import (
	"github.com/elementsproject/glightning/glightning"
	"log"
)

// TODO: listen for `channel_state_changed` and `channel_opened`
// 		so we don't have to refresh peer list every time
// TODO: listen for `shutdown`
// TODO: listen for `sendpay_success` and `sendpay_failure`
// 		so we can update our graph accordingly

func registerSubscriptions(p *glightning.Plugin) {
	p.SubscribeSendPayFailure(OnSendPayFailure)
}

func OnSendPayFailure(sf *glightning.SendPayFailure) {
	log.Printf("send pay failure: %+v\n", sf.Data)
}
