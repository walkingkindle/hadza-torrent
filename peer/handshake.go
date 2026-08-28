package peer

type Handshake struct {
	Reserved [8]byte
	InfoHash [20]byte
	PeerID   [20]byte
}
