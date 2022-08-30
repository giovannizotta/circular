package parallel

import (
	"circular/graph"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"github.com/gammazero/deque"
)

func (r *AbstractRebalance) FindCandidates(exclude string) error {
	r.Node.PeersLock.RLock()
	defer r.Node.PeersLock.RUnlock()

	r.Node.Logln(glightning.Debug, "Looking for candidates")
	peers := r.GetCandidatesList()

	r.Candidates = deque.New[*graph.Channel]()
	for _, p := range peers {
		if p.Id == exclude {
			continue
		}

		for _, peerChannel := range p.Channels {
			// let's see if this channel is a candidate
			if r.IsGoodCandidate(peerChannel) {
				direction := r.GetCandidateDirection(p.Id)
				candidate, err := r.Node.GetGraphChannelFromPeerChannel(peerChannel, direction)
				if err != nil {
					r.Node.Logln(glightning.Unusual, err)
					continue
				}

				r.Node.Logln(glightning.Debug, "adding candidate to candidates:", candidate.ShortChannelId)
				r.Candidates.PushBack(candidate)
			}
		}
	}
	if r.Candidates.Len() == 0 {
		return util.ErrNoCandidates
	}

	r.Node.Logln(glightning.Info, "found ", r.Candidates.Len(), " candidates")
	return nil
}

func (r *AbstractRebalance) GetCandidatesList() []*glightning.Peer {
	if r.CandidatesList == nil {
		// if no CandidatesList was supplied, consider all peers as potential candidates
		return util.GetMapValues(r.Node.Peers)
	} else {
		// if a CandidatesList was supplied, consider only the peers in the CandidatesList as potential candidates
		result := make([]*glightning.Peer, 0)
		for _, peer := range r.CandidatesList {
			if _, ok := r.Node.Peers[peer]; ok {
				result = append(result, r.Node.Peers[peer])
			} else {
				r.Node.Logln(glightning.Unusual, "peer in CandidatesList does not exist: ", peer)
			}
		}

		r.Node.Logln(glightning.Debug, "using CandidatesList: ", r.CandidatesList)

		return result
	}
}

func (r *AbstractRebalance) GetNextCandidate() (*graph.Channel, error) {
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
		if err := r.CanUseChannel(peerChannel); err != nil {
			r.Node.Logln(glightning.Debug, "channel not usable:", err)
			continue
		}
		r.Node.Logln(glightning.Debug, "channel usable")
		return candidate, nil
	}
	return nil, util.ErrNoCandidates
}

func (r *AbstractRebalance) FireCandidates() {
	r.AmountLock.Lock()
	defer r.AmountLock.Unlock()

	carryOn := r.AmountRebalanced+r.InFlightAmount < r.amount
	splitsInFlight := int(r.InFlightAmount / r.splitAmount)

	r.Node.Logln(glightning.Debug, "Firing candidates")
	r.Node.Logln(glightning.Debug, "AmountRebalanced: ", r.AmountRebalanced, ", InFlightAmount: ", r.InFlightAmount, ", Total amount:", r.amount)
	r.Node.Logln(glightning.Debug, "Carry on: ", carryOn, ", splits in flight: ", splitsInFlight)
	for carryOn && splitsInFlight < r.splits {
		candidate, err := r.GetNextCandidate()
		if err != nil {
			// no candidate left
			r.Node.Logln(glightning.Debug, err)
			break
		}
		r.Fire(candidate)

		r.InFlightAmount += r.splitAmount
		carryOn = r.AmountRebalanced+r.InFlightAmount < r.amount
		splitsInFlight = int(r.InFlightAmount / r.splitAmount)

		r.Node.Logln(glightning.Debug, "Carry on: ", carryOn, ", splits in flight: ", splitsInFlight)
	}
}
