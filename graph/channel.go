package graph

import (
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"strconv"
	"strings"
	"time"
)

type Channel struct {
	*glightning.Channel `json:"channel"`
	Liquidity           uint64 `json:"liquidity"`
	Timestamp           int64  `json:"timestamp"`
}

func NewChannel(channel *glightning.Channel, liquidity uint64, timestamp int64) *Channel {
	return &Channel{
		Channel:   channel,
		Liquidity: liquidity,
		Timestamp: timestamp,
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
	minHtlcMsat, _ := strconv.ParseUint(strings.TrimSuffix(c.HtlcMinimumMilliSatoshis, "msat"), 10, 64)
	conditions := []bool{
		c.Liquidity >= amount,
		c.IsActive,
		maxHtlcMsat >= amount,
		minHtlcMsat <= amount,
	}
	return util.All(conditions)
}

func (c *Channel) ResetLiquidity() {
	c.Liquidity = uint64(0.5 * float64(c.Satoshis*1000))
	c.Timestamp = time.Now().Unix()
}
