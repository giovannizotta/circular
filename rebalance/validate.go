package rebalance

import (
	"errors"
	"github.com/elementsproject/glightning/glightning"
	"log"
)

func (r *Rebalance) validatePeerParameters() error {
	//validate that the nodes are not self
	if r.In == r.Self.Id || r.Out == r.Self.Id {
		return errors.New("one of the nodes is self")
	}
	//validate that the nodes are not the same
	if r.In == r.Out {
		return errors.New("incoming and outgoing nodes are the same")
	}
	if len(r.Self.Peers) == 0 {
		return errors.New("no peers yet")
	}
	//validate that In and Out are peers
	if !r.Self.HasPeer(r.In) {
		return errors.New("incoming node is not a peer")
	}
	if !r.Self.HasPeer(r.Out) {
		return errors.New("outgoing node is not a peer")
	}
	return nil
}

func (r *Rebalance) validateLiquidityParameters() error {
	inChannel := r.Self.GetBestPeerChannel(r.In, func(channel *glightning.PeerChannel) uint64 {
		return channel.ReceivableMilliSatoshi
	})
	outChannel := r.Self.GetBestPeerChannel(r.Out, func(channel *glightning.PeerChannel) uint64 {
		return channel.SpendableMilliSatoshi
	})
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

func (r *Rebalance) validateParameters() error {
	if r.In == "" || r.Out == "" {
		return errors.New("missing required parameter: in and out nodes have to be provided")
	}
	if r.Amount == 0 {
		r.Amount = DEFAULT_AMOUNT
		log.Println("amount not provided, using default value", r.Amount)
	}
	if r.MaxPPM == 0 {
		r.MaxPPM = DEFAULT_PPM
		log.Println("maxPPM not provided, using default value", r.MaxPPM)
	}
	err := r.validatePeerParameters()
	if err != nil {
		return err
	}

	err = r.validateLiquidityParameters()
	if err != nil {
		return err
	}

	return nil
}
