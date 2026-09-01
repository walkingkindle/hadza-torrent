// Package tracker is responsible for creating the connection to trackers.
package tracker

import (
	"errors"

	"torrent-client-go/peer"
)

func GetPeers(request peer.AnnounceRequest) (peer.TrackersResponse, error) {
	for _, trackerURL := range request.Trackers {
		url, err := buildTrackerURL(request, trackerURL)
		if err != nil {
			return peer.TrackersResponse{}, errors.New("failed parsing url")
		}
		response, err := getResponse(url)
		if err != nil {
			continue
		}

		if response.Peers != nil {
			return response, nil
		}
	}
	return peer.TrackersResponse{}, errors.New("Could not fetch a response from any of the trackers")
}
