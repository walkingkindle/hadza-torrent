package peer

type Handshake struct {
	InfoHash [20]byte
	PeerID   [20]byte
}
