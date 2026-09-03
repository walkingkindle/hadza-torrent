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
			slog.Error(
				"failed to create tracker client",
				"trackerURL", trackerURL,
				"error", err,
			)
			continue
		}

		slog.Info("found tracker client", "tracker", trackerURL)

		response, err := client.Announce(ctx, request)
		if err != nil {
			slog.Error(
				"failed to fetch response from tracker",
				"trackerURL", trackerURL,
				"error", err,
			)
			continue
		}

		if len(response.Peers) > 0 {
			return response, nil
		}
	}

	return nil, errors.New("could not fetch a response from any of the trackers")
}
