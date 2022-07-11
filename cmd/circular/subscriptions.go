package main

import (
	"circular/node"
	"github.com/elementsproject/glightning/glightning"
)

// TODO: listen for `channel_state_changed` and `channel_opened`
// 		so we don't have to refresh peer list every time
// TODO: listen for `shutdown`

func registerSubscriptions(p *glightning.Plugin) {
	p.SubscribeSendPayFailure(OnSendPayFailure)
	p.SubscribeSendPaySuccess(OnSendPaySuccess)
}

func OnSendPayFailure(sf *glightning.SendPayFailure) {
	node.GetNode().OnPaymentFailure(sf)
}

func OnSendPaySuccess(ss *glightning.SendPaySuccess) {
	node.GetNode().OnPaymentSuccess(ss)
}
