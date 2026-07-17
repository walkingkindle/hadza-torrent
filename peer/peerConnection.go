package peer

import "net"

type PeerConnection struct {
	Conn     net.Conn
	Peer     Peer
	InfoHash [20]byte
	PeerID   [20]byte
}
