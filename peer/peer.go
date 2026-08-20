package peer

import (
	"net"
	"sync/atomic"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

type TrackersResponse struct {
	Peers    []Peer
	Interval int
}

type DownloadState struct {
	Downloaded int64
	Uploaded   int64
	Left       int64
}

type Progress struct {
	uploaded   atomic.Int64
	downloaded atomic.Int64
}

func (p *Progress) DownloadState(totalLength int64) DownloadState {
	downloaded := p.downloaded.Load()

	uploaded := p.uploaded.Load()

	left := max(totalLength-downloaded, 0)

	return DownloadState{
		Uploaded:   uploaded,
		Downloaded: downloaded,
		Left:       left,
	}
}
