package node

import (
	"circular/graph"
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
)

func (n *Node) GetBestPeerChannel(id string, metric func(*glightning.PeerChannel) uint64) *glightning.PeerChannel {
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
	_, ok := n.Peers[id]
	return ok
}

func (n *Node) GetChannelPeerFromScid(scid string) (*glightning.Peer, error) {
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
	if _, ok := n.Graph.Channels[channelId]; !ok {
		return nil, util.ErrNoChannel
	}
	return n.Graph.Channels[channelId], nil
}

func (n *Node) GetOutgoingChannelFromScid(scid string) (*graph.Channel, error) {
	peer, err := n.GetChannelPeerFromScid(scid)
	if err != nil {
		return nil, err
	}

	channelId := scid + "/" + util.GetDirection(n.Id, peer.Id)
	if _, ok := n.Graph.Channels[channelId]; !ok {
		return nil, util.ErrNoOutgoingChannel
	}
	return n.Graph.Channels[channelId], nil
}

func (n *Node) GetIncomingChannelFromScid(scid string) (*graph.Channel, error) {
	peer, err := n.GetChannelPeerFromScid(scid)
	if err != nil {
		return nil, err
	}

	channelId := scid + "/" + util.GetDirection(peer.Id, n.Id)
	if _, ok := n.Graph.Channels[channelId]; !ok {
		return nil, util.ErrNoIncomingChannel
	}
	return n.Graph.Channels[channelId], nil
}

func (n *Node) UpdateChannelBalance(outPeer, scid string, amount uint64) {
	for _, channel := range n.Peers[outPeer].Channels {
		if channel.ShortChannelId == scid {
			channel.SpendableMilliSatoshi -= amount * 1000
			channel.ReceivableMilliSatoshi += amount * 1000
			return
		}
	}
}
