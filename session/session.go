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
	"torrent-client-go/peer-downloader"

	// "torrent-client-go/magnet-parser"
	"torrent-client-go/peer"
	torrentparser "torrent-client-go/torrent"
	"torrent-client-go/types"
)

func DownloadFileFromTorrent(location string) error {
	torrent, err := torrentparser.ParseTorrentFile(location)
	if err != nil {
		return err
	}
	return downloadFileFromTorrent(torrent)
}

// func DonwloadFileFromMagnet(magnetLink string) error {
// 	magnetLink, err := parser.ParseMagnet(magnetLink)
//
// 	if err != nil {
// 		return err
// 	}
//
//
// }

func downloadFileFromTorrent(torrent types.TorrentFile) error {
	file, err := createFile(torrent)
	if err != nil {
		return err
	}
	defer file.Close()

	peerID, err := helpers.GeneratePeerID()
	if err != nil {
		return err
	}

	context := context.Background()
	return downloadLoop(context, torrent, peerID, file)
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
	announceResponse := announcer.PerformAnnounce(
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
