package parallel

import (
	"circular/graph"
	"circular/rebalance"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"github.com/gammazero/deque"
)

/* FindCandidates finds all the candidates for a rebalance going to inPeer that have fees below MaxOutPPM
 * @param inPeer the peer to find candidates for
 * @returns a list of candidates, or an error if none was found
 */
func (r *RebalanceParallel) FindCandidates(inPeer string) error {
	r.Node.PeersLock.RLock()
	defer r.Node.PeersLock.RUnlock()

	r.Candidates = deque.New[*graph.Channel]()
	for _, peer := range r.Node.Peers {
		if peer.Id == inPeer {
			continue
		}

		direction := util.GetDirection(r.Node.Id, peer.Id)
		for _, peerChannel := range peer.Channels {
			channel, err := r.Node.GetGraphChannelFromPeerChannel(peerChannel, direction)
			if err != nil {
				continue
			}
			// let's see if this channel is a candidate
			r.Node.Logln(glightning.Debug, "checking channel:", channel.ShortChannelId)
			r.Node.Logln(glightning.Debug, "feePPM: ", channel.ComputeFeePPM(r.SplitAmount))
			if channel.ComputeFeePPM(r.SplitAmount) < r.MaxOutPPM {
				r.Node.Logln(glightning.Debug, "adding channel to candidates:", channel.ShortChannelId)
				r.Candidates.PushBack(channel)
			}
		}
	}
	if r.Candidates.Len() == 0 {
		return util.ErrNoCandidates
	}

	r.Node.Logln(glightning.Info, "found ", r.Candidates.Len(), " candidates")
	return nil
}

func (r *RebalanceParallel) canUseChannel(channel *glightning.PeerChannel) error {
	// check that the channel is not under the deplete threshold
	depleteAmount := util.Min(r.DepleteUpToAmount,
		uint64(float64(channel.MilliSatoshiTotal)*r.DepleteUpToPercent))
	r.Node.Logln(glightning.Debug, "depleteAmount:", depleteAmount)
	if channel.MilliSatoshiToUs < depleteAmount {
		return util.ErrChannelDepleted
	}

	if channel.State != rebalance.NORMAL {
		return util.ErrChannelNotInNormalState
	}

	if r.Node.IsPeerConnected(channel) == false {
		return util.ErrOutgoingPeerDisconnected
	}

	return nil
}

func (r *RebalanceParallel) GetNextCandidate() (*graph.Channel, error) {
	var candidate *graph.Channel
	r.Node.Logln(glightning.Debug, "getting next candidate")
	for r.Candidates.Len() > 0 {

		r.QueueLock.Lock()
		candidate = r.Candidates.PopFront()
		r.QueueLock.Unlock()
		r.Node.Logln(glightning.Debug, "got candidate:", candidate.ShortChannelId)

		peerChannel, err := r.Node.GetPeerChannelFromGraphChannel(candidate)
		if err != nil {
			r.Node.Logln(glightning.Unusual, "error getting peer channel from graph channel:", err)
			continue
		}

		// check if we can use the channel
		if err := r.canUseChannel(peerChannel); err != nil {
			r.Node.Logln(glightning.Debug, "channel not usable:", err)
			continue
		}
		r.Node.Logln(glightning.Debug, "channel usable")
		return candidate, nil
	}
	return nil, util.ErrNoCandidates
}

// put a candidate at the front of the queue
func (r *RebalanceParallel) EnqueueCandidate(scid string) {
	candidate, err := r.Node.GetOutgoingChannelFromScid(scid)
	if err != nil {
		r.Node.Logln(glightning.Debug, err)
		return
	}

	r.QueueLock.Lock()
	r.Candidates.PushFront(candidate)
	r.QueueLock.Unlock()
}
