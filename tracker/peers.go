// Package tracker is responsible for creating the connection to trackers.
package tracker

import (
	"errors"

	"torrent-client-go/peer"
	"torrent-client-go/types"
)

func GetPeers(torrent types.TorrentFile, peerId string, uploaded string, downloaded string) (peer.TrackersResponse, error) {
	url, err := buildTrackerURL(torrent.InfoHash[:],
		torrent.Length, torrent.Announce, peerId, uploaded, downloaded)
	if err != nil {
		return peer.TrackersResponse{}, errors.New("failed parsing url")
	}

	return getResponse(url)
}
