package types

import "net"

type PeerConnection struct {
	conn     net.Conn
	peer     Peer
	infoHash [20]byte
	peerID   [20]byte
}
