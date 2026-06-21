// Package tracker is responsible for creating the connection to trackers.
package tracker

import (
	"errors"

	"torrent-client-go/types"
)

func GetPeers(torrent types.TorrentFile, peerId string) ([]types.Peer, error) {
	url, err := buildTrackerURL(torrent.InfoHash[:],
		torrent.Length, torrent.Announce, peerId)
	if err != nil {
		return []types.Peer{}, errors.New("failed parsing url")
	}

	return getResponse(url)
}
