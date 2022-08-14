package rebalance

import (
	"circular/graph"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
)

const (
	NORMAL           = "CHANNELD_NORMAL"
	DEFAULT_AMOUNT   = 200000000
	DEFAULT_MAXPPM   = 10
	DEFAULT_ATTEMPTS = 1
	DEFAULT_MAXHOPS  = 8
)

func (r *Rebalance) checkConnections(inChannel, outChannel *glightning.PeerChannel) error {
	r.Node.PeersLock.RLock()
	defer r.Node.PeersLock.RUnlock()

	//validate that the channels are in normal state
	if inChannel.State != NORMAL {
		return util.ErrIncomingChannelNotInNormalState
	}
	if outChannel.State != NORMAL {
		return util.ErrOutgoingChannelNotInNormalState
	}

	// validate that the peers are connected
	if !r.Node.IsPeerConnected(inChannel) {
		return util.ErrIncomingPeerDisconnected
	}
	if !r.Node.IsPeerConnected(outChannel) {
		return util.ErrOutgoingPeerDisconnected
	}
	return nil
}

func (r *Rebalance) checkLiquidity(inChannel, outChannel *glightning.PeerChannel) error {
	r.Node.PeersLock.RLock()
	defer r.Node.PeersLock.RUnlock()

	//validate that the amount is less than the liquidity of the channels
	if (inChannel.MilliSatoshiTotal - inChannel.MilliSatoshiToUs) < r.Amount {
		return util.ErrIncomingChannelDepleted
	}
	if (outChannel.MilliSatoshiToUs) < r.Amount {
		return util.ErrOutgoingChannelDepleted
	}
	return nil
}

func (r *Rebalance) validateLiquidityParameters(out, in *graph.Channel) error {
	r.Node.Logln(glightning.Debug, "validating liquidity parameters")

	r.Node.Logln(glightning.Debug, "refreshing in and out channels")
	// TODO: this refreshes fees, but not spendable and receivable amounts
	r.Node.RefreshChannel(r.OutChannel)
	r.Node.RefreshChannel(r.InChannel)

	inChannel, err := r.Node.GetPeerChannelFromGraphChannel(in)
	if err != nil {
		return err
	}
	outChannel, err := r.Node.GetPeerChannelFromGraphChannel(out)
	if err != nil {
		return err
	}

	if err := r.checkConnections(inChannel, outChannel); err != nil {
		return err
	}

	if err := r.checkLiquidity(inChannel, outChannel); err != nil {
		return err
	}

	r.Node.Logln(glightning.Debug, "liquidity parameters validated")
	return nil
}

func (r *Rebalance) setDefaults() {
	//convert to msatoshi
	r.Amount *= 1000
	if r.Amount == 0 {
		r.Amount = DEFAULT_AMOUNT
		r.Node.Logln(glightning.Debug, "amount not provided, using default value", r.Amount)
	}
	if r.MaxPPM == 0 {
		r.MaxPPM = DEFAULT_MAXPPM
		r.Node.Logln(glightning.Debug, "maxPPM not provided, using default value", r.MaxPPM)
	}
	if r.Attempts <= 0 {
		r.Attempts = DEFAULT_ATTEMPTS
		r.Node.Logln(glightning.Debug, "attempts not provided, using default value", r.Attempts)
	}
	if r.MaxHops <= 0 {
		r.MaxHops = DEFAULT_MAXHOPS
		r.Node.Logln(glightning.Debug, "maxHops not provided, using default value", r.MaxHops)
	}
}
