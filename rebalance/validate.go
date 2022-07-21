package rebalance

import (
	"circular/graph"
	"errors"
	"log"
)

func (r *Rebalance) validateLiquidityParameters(out *graph.Channel, in *graph.Channel) error {
	inChannel, err := r.Node.GetPeerChannelFromNodeID(in.ShortChannelId)
	if err != nil {
		return err
	}
	outChannel, err := r.Node.GetPeerChannelFromNodeID(out.ShortChannelId)
	if err != nil {
		return err
	}
	//validate that the channels are in normal state
	if inChannel.State != NORMAL {
		return errors.New("incoming channel is not in normal state")
	}
	if outChannel.State != NORMAL {
		return errors.New("outgoing channel is not in normal state")
	}
	//validate that the amount is less than the liquidity of the channels
	if (inChannel.ReceivableMilliSatoshi) < r.Amount {
		return errors.New("incoming channel has insufficient remote balance")
	}
	if (outChannel.SpendableMilliSatoshi) < r.Amount {
		return errors.New("outgoing channel has insufficient local balance")
	}
	return nil
}

func (r *Rebalance) setDefaultParameters() error {
	//convert to msatoshi
	r.Amount *= 1000

	if r.Amount == 0 {
		r.Amount = DEFAULT_AMOUNT
		log.Println("amount not provided, using default value", r.Amount)
	}
	if r.MaxPPM == 0 {
		r.MaxPPM = DEFAULT_MAXPPM
		log.Println("maxPPM not provided, using default value", r.MaxPPM)
	}
	if r.Attempts <= 0 {
		r.Attempts = DEFAULT_ATTEMPTS
		log.Println("attempts not provided, using default value", r.Attempts)
	}
	if r.MaxHops <= 0 {
		r.MaxHops = DEFAULT_MAXHOPS
		log.Println("maxHops not provided, using default value", r.MaxHops)
	}
	return nil
}
