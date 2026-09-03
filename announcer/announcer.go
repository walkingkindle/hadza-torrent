// Package announcer takes in a context to provide a list of peers
package announcer

import (
	"context"
	"log/slog"
	"time"

	parser "torrent-client-go/magnet-parser"
	"torrent-client-go/peer"
	"torrent-client-go/tracker"
	"torrent-client-go/types"
)

func AnnounceTorrent(
	ctx context.Context,
	torrent types.TorrentInfo,
	peerID string,
	progress func() peer.DownloadState,
) <-chan []peer.Peer {
	return performAnnounce(ctx, func() (*peer.TrackersResponse, error) {
		return tracker.GetPeers(ctx, peer.AnnounceRequest{
			InfoHash: torrent.InfoHash,
			PeerID:   peerID,
			Trackers: []string{torrent.Announce},
			State:    progress,
		})
	})
}

func AnnounceMagnet(
	ctx context.Context,
	magnet parser.MagnetURI,
	peerID string,
) <-chan []peer.Peer {
	return performAnnounce(ctx, func() (*peer.TrackersResponse, error) {
		return tracker.GetPeers(ctx, peer.AnnounceRequest{
			InfoHash: magnet.Infohash,
			PeerID:   peerID,
			Trackers: magnet.Trackers,
			State:    nil,
		})
	})
}

func performAnnounce(
	ctx context.Context,
	announce func() (*peer.TrackersResponse, error),
	// progress func() peer.DownloadState,
) <-chan []peer.Peer {
	peersCh := make(chan []peer.Peer, 1)

	go func() {
		defer close(peersCh)

		const (
			minBackoff  = 15 * time.Second
			maxBackoff  = 15 * time.Minute
			minInterval = 60 * time.Second
			maxInterval = 30 * time.Minute
		)

		backOff := minBackoff
		for {
			response, err := announce()
			if err != nil {
				slog.Warn("announce failed, backing off",
					"err", err, "retry_in", backOff)

				timer := time.NewTimer(backOff)
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
				}

				backOff = min(backOff*2, maxBackoff)
				continue
			}
			backOff = minBackoff

			interval := time.Duration(response.Interval) * time.Second
			interval = max(interval, minInterval)
			interval = min(interval, maxInterval)

			select {
			case <-peersCh:

			default:
			}
			select {
			case peersCh <- response.Peers:
			default:
			}

			timer := time.NewTimer(interval)

			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}

		}
	}()
	return peersCh
}
