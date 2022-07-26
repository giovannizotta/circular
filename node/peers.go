package node

import (
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
	"log"
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

func (n *Node) GetPeerChannelFromNodeID(scid string) (*glightning.PeerChannel, error) {
	for _, peer := range n.Peers {
		for _, channel := range peer.Channels {
			if channel.ShortChannelId == scid {
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

func (n *Node) RefreshPeerChannels(id string) {
	channels, err := n.lightning.ListChannelsBySource(id)
	if err != nil {
		log.Println(err)
	}
	n.Graph.RefreshChannels(channels)
}
