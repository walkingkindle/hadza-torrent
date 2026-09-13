package peer

import (
	"net"
	"strings"
	"sync/atomic"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func (p Peer) String() string {
	return strings.Join([]string{p.IP.To4().String(), string(p.Port)}, ", ")
}

type TrackersResponse struct {
	Peers    []Peer
	Interval int
	Seeders  int
	Leechers int
}

type DownloadState struct {
	Downloaded int64
	Uploaded   int64
	Left       int64
}

type AnnounceRequest struct {
	InfoHash string
	PeerID   string
	State    func() DownloadState
	Trackers []string
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
