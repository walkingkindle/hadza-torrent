// Package tracker is responsible for creating the connection to trackers.
package tracker

import (
	"context"
	"errors"
	"log/slog"

	"torrent-client-go/peer"
)

func GetPeers(ctx context.Context, request peer.AnnounceRequest) (*peer.TrackersResponse, error) {
	for _, trackerURL := range request.Trackers {

		slog.Info("calling tracker", "trackerURL", trackerURL)

		client, err := NewTrackerClient(trackerURL)
		if err != nil {
			return nil, err
		}

		response, err := client.Announce(ctx, request)
		if err != nil {
			slog.Error("Failed to fetch response from the trackerURL", "trackerURL", trackerURL, "error", err.Error())
			return nil, err
		}

		if response.Peers != nil {
			return response, nil
		}
	}
	return nil, errors.New("Could not fetch a response from any of the trackers")
}
