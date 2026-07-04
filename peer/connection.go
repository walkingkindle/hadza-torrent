// Package peer tries to handshake with the return peers and the connection struct
package peer

import (
	"net"
	"time"

	"torrent-client-go/types"
)

func Connect(peer types.Peer, infohash string, peerID [20]byte) (*types.PeerConnection, error) {
	net.DialTimeout("tcp", peer.IP.String(), 2*time.Second)
}
