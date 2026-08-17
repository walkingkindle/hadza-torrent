package peer

import "net"

type Peer struct {
	IP   net.IP
	Port uint16
}

type TrackersResponse struct {
	Peers    []Peer
	Interval int
}
