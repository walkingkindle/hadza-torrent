package discovery

import (
	"context"
	"net"
)

type PeerDiscovery interface {
	Discover(ctx context.Context, infoHash [20]byte) ([]net.TCPAddr, error)
}
