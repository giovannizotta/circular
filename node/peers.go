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

func (n *Node) IsPeerConnected(channel *glightning.PeerChannel) bool {
	peer, err := n.GetChannelPeerFromScid(channel.ShortChannelId)
	if err != nil {
		return false
	}

	return peer.Connected
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

	n.PeersLock.RLock()
	defer n.PeersLock.RUnlock()
	channelId := scid + "/" + util.GetDirection(n.Id, peer.Id)
	channel, err := n.Graph.GetChannel(channelId)
	if err == util.ErrNoChannel {
		return nil, util.ErrNoOutgoingChannel
	}
	return channel, err
}

func (n *Node) GetIncomingChannelFromScid(scid string) (*graph.Channel, error) {
	peer, err := n.GetChannelPeerFromScid(scid)
	if err != nil {
		return nil, err
	}

	n.PeersLock.RLock()
	defer n.PeersLock.RUnlock()

	channelId := scid + "/" + util.GetDirection(peer.Id, n.Id)
	channel, err := n.Graph.GetChannel(channelId)
	if err == util.ErrNoChannel {
		return nil, util.ErrNoIncomingChannel
	}
	return channel, err
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

func (n *Node) OnConnect(c *glightning.ConnectEvent) {
	n.PeersLock.Lock()
	defer n.PeersLock.Unlock()

	if _, ok := n.Peers[c.PeerId]; ok {
		n.Peers[c.PeerId].Connected = true
	}
}

func (n *Node) OnDisconnect(c *glightning.DisconnectEvent) {
	n.PeersLock.Lock()
	defer n.PeersLock.Unlock()

	if _, ok := n.Peers[c.PeerId]; ok {
		n.Peers[c.PeerId].Connected = false
	}
}
