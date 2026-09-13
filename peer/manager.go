package peer

type PeerManager struct {
	peers map[string]Peer
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
		peers: make(map[string]Peer),
	}
}

func (m *PeerManager) AddPeers(peers []Peer) {
	for _, p := range peers {
		m.peers[p.String()] = p
	}
}
