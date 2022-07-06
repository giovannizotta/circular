package main

import (
	"github.com/elementsproject/glightning/glightning"
	"log"
)

func registerSubscriptions(p *glightning.Plugin) {
	//tmp
	p.SubscribeChannelOpened(OnChannelOpened)
}

func OnChannelOpened(co *glightning.ChannelOpened) {
	//tmp
	log.Printf("channel opened with %s for %s. is locked? %v", co.PeerId, co.FundingSatoshis, co.FundingLocked)
}
