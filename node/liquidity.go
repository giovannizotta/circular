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

		n.Logf(glightning.Debug, "channel %s failed, opposite channel is %s", channelId, oppositeChannelId)

		n.Graph.UpdateChannel(channelId, oppositeChannelId, update.Amount)
	}
}
