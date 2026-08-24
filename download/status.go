package download

import (
	"sync"

	"torrent-client-go/peer"
)

type DownloadStatus struct {
	Mu         sync.Mutex
	Done       []bool
	InProgress []bool
}

func (s *DownloadStatus) ClaimPiece(conn *peer.PeerConnection) (int, bool) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for i := range s.Done {
		if s.Done[i] || s.InProgress[i] {
			continue
		}

		if !conn.HasPiece(i) {
			continue
		}

		s.InProgress[i] = true
		return i, true
	}

	return 0, false
}

func (s *DownloadStatus) Complete() bool {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	for _, Done := range s.Done {
		if !Done {
			return false
		}
	}

	return true
}

func (s *DownloadStatus) CompletePiece(i int) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.InProgress[i] = false
	s.Done[i] = true
}

func (s *DownloadStatus) ReleasePiece(i int) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.InProgress[i] = false
}
