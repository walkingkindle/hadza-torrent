// Package session owns the download end-to-end
package session

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"torrent-client-go/download"
	"torrent-client-go/helpers"
	"torrent-client-go/peer"
	torrentparser "torrent-client-go/torrent"
	"torrent-client-go/types"
)

func DownloadFileFromTorrent(location string) error {
	torrent, err := torrentparser.ParseTorrentFile(location)
	if err != nil {
		return err
	}

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
	err = downloadLoop(context, torrent, peerID, file)
	if err != nil {
		return err
	}

	return nil
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
	announceResponse := performAnnounce(
		ctx,
		torrent,
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
				connection, err := peer.Connect(p, torrent.InfoHash, peerID)
				if err != nil {
					slog.Warn("peer rejected connection", "peer", p, "error", err)
					return
				}
				if err = download.Download(&connection, torrent, file, state); err != nil {
					slog.Warn("peer download failed", "peer", p.IP, "error", err)
					return
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

// for _, peer := range trackersResponse.Peers {
// err = retry.New(
// 	retry.Attempts(0),
// 	retry.Delay(2*time.Second),
// 	retry.MaxDelay(2*time.Minute),
// 	retry.LastErrorOnly(true),
// 	retry.RetryIf(func(err error) bool {
// 		if errors.Is(err, download.ErrLocal) {
// 			return false
// 		}
// 		return noProgress < 5
// 	}),
// ).Do(func() error {
// 	conn, err := connectToPeer(peer, torrent, peerID)
// 	if err != nil {
// 		noProgress++
// 		return err
// 	}
// 	before := helpers.CountTrue(done)
// 	err = download.Download(conn, torrent, file, done)
//
// 	if before == helpers.CountTrue(done) {
// 		noProgress++
// 	} else {
// 		noProgress = 0
// 	}
//
// 	if helpers.CountTrue(done) == len(done) {
// 		return nil
// 	}
//
// 	return err
// })
//
// // 4. If we timed out, loop again to re-announce
// if helpers.CountTrue(done) == len(done) {
// } else {
// 	slog.Warn("Finding peer failed,", "err", err)
// }
