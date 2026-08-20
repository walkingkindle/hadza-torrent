// Package session owns the download end-to-end
package session

import (
	"context"
	"os"

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

	peerId, err := helpers.GeneratePeerID()
	if err != nil {
		return err
	}

	done := make([]bool, len(torrent.PieceHashes))
	context := context.Background()
	err = downloadLoop(context, torrent, peerId, done, file)
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

func downloadLoop(ctx context.Context, torrent types.TorrentFile, peerID string, done []bool, file *os.File) error {
	p := &peer.Progress{}
	_ = performAnnounce(
		ctx,
		torrent,
		peerID,
		func() peer.DownloadState {
			return p.DownloadState(int64(torrent.Length))
		},
	)

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
