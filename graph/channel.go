package graph

import (
	"github.com/elementsproject/glightning/glightning"
	"time"
)

type Channel struct {
	*glightning.Channel `json:"channel"`
	Liquidity           uint64 `json:"liquidity"`
	Timestamp           int64  `json:"timestamp"`
	maxHtlcMsat         uint64
	minHtlcMsat         uint64
}

func NewChannel(channel *glightning.Channel, liquidity uint64, timestamp int64) *Channel {
	return &Channel{
		Channel:     channel,
		Liquidity:   liquidity,
		Timestamp:   timestamp,
		maxHtlcMsat: channel.HtlcMaximumMilliSatoshis.MSat(),
		minHtlcMsat: channel.HtlcMinimumMilliSatoshis.MSat(),
	}
}

func (c *Channel) ComputeFee(amount uint64) uint64 {
	result := c.BaseFeeMillisatoshi
	// get the ceiling of the integer division
	numerator := (amount / 1000) * c.FeePerMillionth
	var proportionalFee uint64 = 0
	if numerator > 0 {
		proportionalFee = ((numerator - 1) / 1000) + 1
	}
	result += proportionalFee
	return result
}

func (c *Channel) ComputeFeePPM(amount uint64) uint64 {
	return c.ComputeFee(amount) * 1000000 / amount
}

func (c *Channel) GetHop(amount uint64, delay uint32) glightning.RouteHop {
	return glightning.RouteHop{
		Id:             c.Destination,
		ShortChannelId: c.ShortChannelId,
		AmountMsat:     glightning.AmountFromMSat(amount),
		Delay:          delay,
		Direction:      c.GetDirection(),
	}
}

func (c *Channel) GetDirection() uint32 {
	if c.Source < c.Destination {
		return 0
	}
	return 1
}

func (c *Channel) CanForward(amount uint64) bool {
	return c.IsActive &&
		c.Liquidity >= amount &&
		c.maxHtlcMsat >= amount &&
		c.minHtlcMsat <= amount
}

func (c *Channel) ResetLiquidity() {
	c.Liquidity = uint64(0.5 * float64(c.AmountMsat.MSat()))
	c.Timestamp = time.Now().Unix()
}
