// Package peerdownloader takes one peer and tries to download file from it
package peerdownloader

import (
	"context"
	"log/slog"
	"os"

	"torrent-client-go/download"
	"torrent-client-go/peer"
	"torrent-client-go/types"
)

func DownloadFromPeer(downloadCtx context.Context, p peer.Peer, torrent types.TorrentFile, peerID [20]byte, file *os.File, state *download.DownloadStatus) error {
	connection, err := peer.Connect(p, torrent.InfoHash, peerID)
	if err != nil {
		slog.Warn("peer rejected connection", "peer", p, "error", err)
		return err
	}

	err = connection.SendInterested()
	if err != nil {
		slog.Error("peer connection succeeeded but failed to send interested")
		connection.Conn.Close()
		return err
	}
	if err = download.Download(downloadCtx, &connection, torrent, file, state); err != nil {
		return err
	}
	return nil
}
