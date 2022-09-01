package util

import (
	"errors"
	"fmt"
)

type ErrRouteTooExpensive struct {
	FeePPM uint64
	MaxPPM uint64
}

func NewRouteTooExpensiveError(feePPM uint64, maxPPM uint64) ErrRouteTooExpensive {
	return ErrRouteTooExpensive{
		FeePPM: feePPM,
		MaxPPM: maxPPM,
	}
}

func (e ErrRouteTooExpensive) Error() string {
	return fmt.Sprintf("route too expensive. Cheapest route found was %d ppm, but maxppm is %d", e.FeePPM, e.MaxPPM)
}

var (
	ErrSendPayTimeout      = errors.New("200:Timed out while waiting")
	ErrTemporaryFailure    = errors.New("204:failed: WIRE_TEMPORARY_CHANNEL_FAILURE (reply from remote)")
	ErrWireFeeInsufficient = errors.New("204:failed: WIRE_FEE_INSUFFICIENT (reply from remote)")

	ErrNoRequiredParameter         = errors.New("missing required parameter")
	ErrSelfNode                    = errors.New("one of the nodes is self")
	ErrNoPeers                     = errors.New("no peers yet")
	ErrSameIncomingAndOutgoingNode = errors.New("incoming and outgoing nodes are the same")
	ErrNoPeerChannel               = errors.New("not a peer or peer channel")
	ErrNoSuchNode                  = errors.New("no such node")
	ErrNoPeer                      = errors.New("no peer")
	ErrFirstPeerNotReady           = errors.New("first peer not ready")
	ErrCircularStopped             = errors.New("circular has been stopped. Use 'circular-resume' to resume activity")

	ErrNoGraphToLoad = errors.New("no graph to load")
	ErrNoRoute       = errors.New("no route")

	ErrAmountLessThanSplitAmount      = errors.New("amount is less than split amount")
	ErrAmountNotMultipleOfSplitAmount = errors.New("amount is not a multiple of split amount")
	ErrDepleteUpToPercentInvalid      = errors.New("deplete up to percent invalid, it must be between 0 and 1")

	ErrNoChannel               = errors.New("no channel")
	ErrNoCandidates            = errors.New("no candidates")
	ErrChannelDepleted         = errors.New("channel is depleted")
	ErrChannelFilled           = errors.New("channel is filled")
	ErrIncomingChannelDepleted = errors.New("incoming channel does not have enough remote balance")
	ErrOutgoingChannelDepleted = errors.New("outgoing channel does not have enough local balance")

	ErrNoOutgoingChannel               = errors.New("no outgoing channel")
	ErrNoIncomingChannel               = errors.New("no incoming channel")
	ErrChannelNotInNormalState         = errors.New("channel is not in normal state")
	ErrOutgoingChannelNotInNormalState = errors.New("outgoing channel is not in normal state")
	ErrIncomingChannelNotInNormalState = errors.New("incoming channel is not in normal state")

	ErrChannelNotFound         = errors.New("channel not found")
	ErrOppositeChannelNotFound = errors.New("opposite channel not found")

	ErrIncomingPeerDisconnected = errors.New("incoming peer is disconnected")
	ErrOutgoingPeerDisconnected = errors.New("outgoing peer is disconnected")
)
