package graph

import (
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"strconv"
	"strings"
)

type Channel struct {
	*glightning.Channel `json:"channel"`
	Liquidity           uint64 `json:"liquidity"`
}

func NewChannel(channel *glightning.Channel, liquidity uint64) *Channel {
	return &Channel{
		Channel:   channel,
		Liquidity: liquidity,
	}
}

func (c *Channel) ComputeFee(amount uint64) uint64 {
	result := c.BaseFeeMillisatoshi
	// get the ceiling of the integer division
	numerator := (amount / 1000) * c.FeePerMillionth
	proportionalFee := ((numerator - 1) / 1000) + 1
	result += proportionalFee
	return result
}

func (c *Channel) GetHop(amount uint64, delay uint) glightning.RouteHop {
	return glightning.RouteHop{
		Id:             c.Destination,
		ShortChannelId: c.ShortChannelId,
		MilliSatoshi:   amount,
		Delay:          delay,
		Direction:      c.GetDirection(),
	}
}

func (c *Channel) GetDirection() uint8 {
	if c.Source < c.Destination {
		return 0
	}
	return 1
}

func (c *Channel) CanForward(amount uint64) bool {
	maxHtlcMsat, _ := strconv.ParseUint(strings.TrimSuffix(c.HtlcMaximumMilliSatoshis, "msat"), 10, 64)
	conditions := []bool{
		c.Liquidity >= amount,
		c.IsActive,
		maxHtlcMsat >= amount,
	}
	return util.All(conditions)
}
