package node

import (
	"github.com/elementsproject/glightning/glightning"
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
		n.Logf(glightning.Debug, "LiquidityUpdate: %+v", update)

		direction := strconv.Itoa(update.Direction)
		channelId := update.ShortChannelID + "/" + direction

		oppositeDirection := strconv.Itoa(update.Direction ^ 0x1)
		oppositeChannelId := update.ShortChannelID + "/" + oppositeDirection

		n.Logln(glightning.Debug, "failed from "+n.Graph.Channels[channelId].Source+" to "+n.Graph.Channels[channelId].Destination)
		n.Logf(glightning.Debug, "channel %s failed, opposite channel is %s", channelId, oppositeChannelId)

		n.graphLock.L.Lock()
		if _, ok := n.Graph.Channels[channelId]; ok {
			n.Graph.Channels[channelId].Liquidity = update.Amount
		} else {
			n.Logln(glightning.Unusual, "channel not found:", channelId)
		}

		if _, ok := n.Graph.Channels[oppositeChannelId]; ok {
			n.Graph.Channels[oppositeChannelId].Liquidity =
				n.Graph.Channels[oppositeChannelId].Satoshis*1000 - update.Amount
		} else {
			n.Logln(glightning.Unusual, "opposite channel not found:", oppositeChannelId)
		}
		n.graphLock.L.Unlock()
	}
}
