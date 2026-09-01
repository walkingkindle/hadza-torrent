// Package session owns the download end-to-end
package session

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"torrent-client-go/announcer"
	"torrent-client-go/download"
	"torrent-client-go/helpers"
	parser "torrent-client-go/magnet-parser"
	"torrent-client-go/metadata"
	"torrent-client-go/peer"
	"torrent-client-go/peer-downloader"
	torrentparser "torrent-client-go/torrent"
	"torrent-client-go/types"
)

func DownloadFile(location string) error {
	context := context.Background()
	peerID, err := helpers.GeneratePeerID()
	if err != nil {
		return err
	}
	if parser.IsAMagnet(location) {
		return downloadFileFromMagnet(context, location, peerID)
	}

	torrent, err := torrentparser.ParseTorrentFile(location)
	if err != nil {
		return err
	}

	return downloadFileFromTorrent(context, torrent, peerID)
}

func downloadFileFromMagnet(ctx context.Context, magnetLink string, peerID [20]byte) error {
	magnetURI, err := parser.ParseMagnet(magnetLink)
	if err != nil {
		return err
	}

	fmt.Printf("%# v\n", magnetURI)

	torrent, err := metadata.Fetch(ctx, magnetURI, peerID)
	if err != nil {
		return err
	}

	return downloadFileFromTorrent(ctx, torrent, peerID)
}

func downloadFileFromTorrent(ctx context.Context, torrent types.TorrentFile, peerID [20]byte) error {
	// TODO: Support multiple files here
	file, err := createFile(torrent)
	if err != nil {
		return err
	}
	defer file.Close()

	return downloadLoop(ctx, torrent, peerID, file)
}

func createFile(torrent types.TorrentFile) (*os.File, error) {
	file, err := os.OpenFile(torrent.Name, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}

	if err := file.Truncate(int64(torrent.Length)); err != nil {
		return nil, err
	}
	return file, nil
}

func downloadLoop(ctx context.Context, torrent types.TorrentFile, peerID [20]byte, file *os.File) error {
	downloadCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	p := &peer.Progress{}
	announceResponse := announcer.AnnounceTorrent(
		downloadCtx,
		types.TorrentInfo{InfoHash: string(torrent.InfoHash[:]), Announce: torrent.Announce, Length: int64(torrent.Length)},
		string(peerID[:]),
		func() peer.DownloadState {
			return p.DownloadState(int64(torrent.Length))
		},
	)
	state := &download.DownloadStatus{
		Done:       make([]bool, len(torrent.PieceHashes)),
		InProgress: make([]bool, len(torrent.PieceHashes)),
	}
	var wg sync.WaitGroup

	for response := range announceResponse {
		if state.Complete() {
			break
		}
		for _, p := range response {
			if state.Complete() {
				cancel()
				break
			}
			wg.Add(1)
			go func(p peer.Peer) {
				defer wg.Done()
				if err := peerdownloader.DownloadFromPeer(downloadCtx, p, torrent, peerID, file, state); err != nil {
					slog.Warn("peer download failed", "peer", p.IP, "error", err)
				}

				if state.Complete() {
					cancel()
				}
			}(p)
		}
	}
	wg.Wait()

	if !state.Complete() {
		if err := downloadCtx.Err(); err != nil {
			return err
		}

		return fmt.Errorf("download incomplete got %d of %d pieces",
			state.DoneCount(), state.TotalPieces())
	}

	slog.Info("torrent download complete",
		"file",
		torrent.Name,
		"pieces", state.TotalPieces())

	return nil
}
