package node

import (
	"circular/util"
	"github.com/elementsproject/glightning/glightning"
)

func (s *Node) GetBestPeerChannel(id string, metric func(*glightning.PeerChannel) uint64) *glightning.PeerChannel {
	channels := s.Peers[id].Channels
	best := channels[0]
	for _, channel := range channels {
		if metric(channel) > metric(best) {
			best = channel
		}
	}
	return best
}

func (s *Node) GetPeerChannelFromNodeID(scid string) (*glightning.PeerChannel, error) {
	for _, peer := range s.Peers {
		for _, channel := range peer.Channels {
			if channel.ShortChannelId == scid {
				return channel, nil
			}
		}
	}
	return nil, util.ErrNoPeerChannel
}

func (s *Node) HasPeer(id string) bool {
	_, ok := s.Peers[id]
	return ok
}

func (s *Node) GetChannelPeerFromScid(scid string) (*glightning.Peer, error) {
	for _, peer := range s.Peers {
		for _, channel := range peer.Channels {
			if channel.ShortChannelId == scid {
				return peer, nil
			}
		}
	}
	return nil, util.ErrNoPeerChannel
}
