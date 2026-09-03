package tracker

import (
	"context"
	"fmt"
	"net/url"

	"torrent-client-go/peer"
)

type TrackerClient interface {
	Announce(ctx context.Context, req peer.AnnounceRequest) (*peer.TrackersResponse, error)
}

func NewTrackerClient(trackerURL string) (TrackerClient, error) {
	u, err := url.Parse(trackerURL)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https":
		return &HTTPTrackerClient{BaseURL: u}, nil
	case "udp":
		return &UDPTrackerClient{BaseURL: u}, nil
	// case "ws", "wss":
	// 	return &WSSTrackerClient{URL: u}, nil
	default:
		return nil, fmt.Errorf("unsupported tracker scheme: %s", u.Scheme)
	}
}
