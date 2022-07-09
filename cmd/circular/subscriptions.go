package main

import (
	"github.com/elementsproject/glightning/glightning"
	"log"
)

// TODO: listen for `channel_state_changed` and `channel_opened`
// 		so we don't have to refresh peer list every time
// TODO: listen for `shutdown`

func registerSubscriptions(p *glightning.Plugin) {
	p.SubscribeSendPayFailure(OnSendPayFailure)
	p.SubscribeSendPaySuccess(OnSendPaySuccess)
}

func OnSendPayFailure(sf *glightning.SendPayFailure) {
	log.Printf("send pay failure: %+v\n", sf.Data)
	log.Println("ErringNode: ", sf.Data.ErringNode)
	log.Println("ErringChannel: ", sf.Data.ErringChannel)
	log.Println("ErringIndex: ", sf.Data.ErringIndex)
	log.Println("ErringDirection: ", sf.Data.ErringDirection)
}

func OnSendPaySuccess(ss *glightning.SendPaySuccess) {
	log.Printf("send pay success: %+v\n", ss)
}
