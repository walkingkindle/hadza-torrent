// Package metadata uses the magnet link to fetch more information about the torrent file
package metadata

import (
	"context"
	"errors"
	"strconv"

	"torrent-client-go/announcer"
	"torrent-client-go/magnet-parser"
	"torrent-client-go/peer"
	"torrent-client-go/types"
)

func Fetch(ctx context.Context, magnet parser.MagnetURI, peerID [20]byte) (types.TorrentFile, error) {
	p := &peer.Progress{}
	length, err := strconv.ParseInt(magnet.Length, 10, 64)
	if err != nil {
		return types.TorrentFile{}, err
	}
	announceResponse := announcer.PerformAnnounce(
		ctx,
		types.TorrentInfo{Announce: magnet.Trackers[0], InfoHash: magnet.Infohash, Length: length},
		string(peerID[:]),
		func() peer.DownloadState {
			return p.DownloadState(length)
		},
	)
	for response := range announceResponse {
		for _, p := range response {
			var infohash [20]byte
			copy(infohash[:], []byte(magnet.Infohash))
			conn, err := peer.Connect(p, infohash, peerID)
			if err != nil {
				continue
			}

			torrent, err := fetchFromPeer(ctx, conn)
			if err != nil {
				return types.TorrentFile{}, err
			}

			return torrent, nil
		}
	}
	return types.TorrentFile{}, errors.New("could not find a valid response from any of the peers")
}

func fetchFromPeer(ctx context.Context, conn peer.PeerConnection) (types.TorrentFile, error) {
	panic("unimplemented")
}
