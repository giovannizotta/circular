package node

import (
	"circular/graph"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
)

func (n *Node) GetBestPeerChannel(id string, metric func(*glightning.PeerChannel) uint64) *glightning.PeerChannel {
	n.PeersLock.RLock()
	defer n.PeersLock.RUnlock()

	channels := n.Peers[id].Channels
	best := channels[0]
	for _, channel := range channels {
		if metric(channel) > metric(best) {
			best = channel
		}
	}
	return best
}

func (n *Node) GetPeerChannelFromGraphChannel(graphChannel *graph.Channel) (*glightning.PeerChannel, error) {
	n.PeersLock.RLock()
	defer n.PeersLock.RUnlock()

	for _, peer := range n.Peers {
		for _, channel := range peer.Channels {
			if channel.ShortChannelId == graphChannel.ShortChannelId {
				return channel, nil
			}
		}
	}
	return nil, util.ErrNoPeerChannel
}

func (n *Node) HasPeer(id string) bool {
	n.PeersLock.RLock()
	defer n.PeersLock.RUnlock()

	_, ok := n.Peers[id]
	return ok
}

func (n *Node) GetChannelPeerFromScid(scid string) (*glightning.Peer, error) {
	n.PeersLock.RLock()
	defer n.PeersLock.RUnlock()

	for _, peer := range n.Peers {
		for _, channel := range peer.Channels {
			if channel.ShortChannelId == scid {
				return peer, nil
			}
		}
	}
	return nil, util.ErrNoPeerChannel
}

func (n *Node) GetGraphChannelFromPeerChannel(channel *glightning.PeerChannel, direction string) (*graph.Channel, error) {
	channelId := channel.ShortChannelId + "/" + direction
	return n.Graph.GetChannel(channelId)
}

func (n *Node) GetOutgoingChannelFromScid(scid string) (*graph.Channel, error) {
	peer, err := n.GetChannelPeerFromScid(scid)
	if err != nil {
		return nil, err
	}

	channelId := scid + "/" + util.GetDirection(n.Id, peer.Id)
	return n.Graph.GetChannel(channelId)
}

func (n *Node) GetIncomingChannelFromScid(scid string) (*graph.Channel, error) {
	peer, err := n.GetChannelPeerFromScid(scid)
	if err != nil {
		return nil, err
	}

	channelId := scid + "/" + util.GetDirection(peer.Id, n.Id)
	return n.Graph.GetChannel(channelId)
}

func (n *Node) UpdateChannelBalance(outPeer, scid string, amount uint64) {
	n.PeersLock.Lock()
	defer n.PeersLock.Unlock()

	for _, channel := range n.Peers[outPeer].Channels {
		if channel.ShortChannelId == scid {
			channel.MilliSatoshiToUs -= amount * 1000
			return
		}
	}
}
