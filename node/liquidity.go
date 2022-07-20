package node

import (
	"log"
	"strconv"
)

type LiquidityUpdate struct {
	Amount         uint64
	ShortChannelID string
	Direction      int
}

func (n *Node) UpdateLiquidity() {
	for {
		update := <-n.LiquidityUpdateChan

		direction := strconv.Itoa(update.Direction)
		channelId := update.ShortChannelID + "/" + direction

		oppositeDirection := strconv.Itoa(update.Direction ^ 0x1)
		oppositeChannelId := update.ShortChannelID + "/" + oppositeDirection

		log.Println("failed from " + n.Graph.Channels[channelId].Source + " to " + n.Graph.Channels[channelId].Destination)
		log.Printf("channel %s failed, opposite channel is %s\n", channelId, oppositeChannelId)

		if _, ok := n.Graph.Channels[channelId]; ok {
			n.Graph.Channels[channelId].Liquidity = update.Amount
		} else {
			log.Println("channel not found:", channelId)
		}

		if _, ok := n.Graph.Channels[oppositeChannelId]; ok {
			n.Graph.Channels[oppositeChannelId].Liquidity =
				n.Graph.Channels[oppositeChannelId].Satoshis*1000 - update.Amount
		} else {
			log.Println("opposite channel not found:", oppositeChannelId)
		}
	}
}
