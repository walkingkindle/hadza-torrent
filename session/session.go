// Package session owns the download end-to-end
package session

import (
	"log/slog"
	"os"
	"time"

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

	peerId, err := generatePeerID()
	if err != nil {
		return err
	}

	done := make([]bool, len(torrent.PieceHashes))
	err = downloadLoop(torrent, peerId, done, file)
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

func downloadLoop(torrent types.TorrentFile, peerID [20]byte, done []bool, file *os.File) error {
	const minBackoff = 15 * time.Second
	const maxBackoff = 15 * time.Minute
	backoff := minBackoff
	for {
		trackersResponse, err := tracker.GetPeers(torrent, string(peerID[:]), "0", "0")
		if err != nil {
			slog.Warn("announce failed, backing off",
				"err", err, "retry_in", backoff)

			time.Sleep(backoff)
			backoff = min(backoff*2, maxBackoff)

			continue
		}

		backoff = minBackoff

		timer := time.NewTimer(time.Second * time.Duration(trackersResponse.Interval))

		timedOut := false

		for _, peer := range trackersResponse.Peers {
			select {
			case <-timer.C:
				timedOut = true
			default:
				// Timer not up yet, continue with peer
			}

			if timedOut {
				break
			}

			noProgress := 0
			err = retry.New(
				retry.Attempts(0),
				retry.Delay(2*time.Second),
				retry.MaxDelay(2*time.Minute),
				retry.LastErrorOnly(true),
				retry.RetryIf(func(err error) bool {
					if errors.Is(err, download.ErrLocal) {
						return false
					}
					return noProgress < 5
				}),
			).Do(func() error {
				conn, err := connectToPeer(peer, torrent, peerID)
				if err != nil {
					noProgress++
					return err
				}
				before := helpers.CountTrue(done)
				err = download.Download(conn, torrent, file, done)

				if before == helpers.CountTrue(done) {
					noProgress++
				} else {
					noProgress = 0
				}

				if helpers.CountTrue(done) == len(done) {
					return nil
				}

				return err
			})
		}

		// 4. If we timed out, loop again to re-announce
		if timedOut {
			fmt.Println("Interval reached, re-announcing...")
			continue // Restart the outer loop to call GetPeers again
		}

		if helpers.CountTrue(done) == len(done) {
			break
		} else {
			slog.Warn("Finding peer failed,", "err", err)
			continue
		}
	}
	return nil
}

func generatePeerID() ([20]byte, error) {
	var bytes [20]byte
	_, err := rand.Read(bytes[:])
	if err != nil {
		return [20]byte{}, err
	}

	return bytes, nil
}
