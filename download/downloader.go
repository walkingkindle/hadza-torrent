// Package download handles the download feed and writes to file
package download

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"torrent-client-go/peer"
	"torrent-client-go/types"
)

var ErrLocal = errors.New("local failure")

func (s *DownloadStatus) DoneCount() int {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	count := 0
	for _, Done := range s.Done {
		if Done {
			count++
		}
	}

	return count
}

func (s *DownloadStatus) TotalPieces() int {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	return len(s.Done)
}

func Download(ctx context.Context, conn *peer.PeerConnection, torrent types.TorrentFile, file *os.File, downloadStatus *DownloadStatus) error {
	stop := context.AfterFunc(ctx, func() {
		conn.Conn.Close()
	})
	defer stop()
	conn.Conn.SetDeadline(time.Now().Add(60 * time.Second))

	defer conn.Conn.Close()
	err := conn.WaitForUnchoke()
	if err != nil {
		return err
	}

	slog.Info("starting download",
		"file", torrent.Name,
		"bytes", torrent.Length,
		"pieces", len(torrent.PieceHashes),
		"piece_length", torrent.PieceLength)

	started := time.Now()

	for !downloadStatus.Complete() {
		i, ok := downloadStatus.ClaimPiece(conn)
		if !ok {
			slog.Info("waiting for the peer to offer a piece we still need",
				"peer", conn.Peer.IP, "advertised", conn.PieceCount(), "have", downloadStatus.DoneCount())

			if err := awaitAdvertisement(conn); err != nil {
				return fmt.Errorf("got %d of %d pieces from %s before it stopped offering any: %w",
					downloadStatus.DoneCount(), downloadStatus.TotalPieces(), conn.Peer.IP, err)
			}
			continue
		}

		piece, err := downloadPiece(ctx, i, torrent, conn)
		if err != nil {
			downloadStatus.ReleasePiece(i)
			return err
		}

		if _, err := file.WriteAt(piece, int64(i*torrent.PieceLength)); err != nil {
			downloadStatus.ReleasePiece(i)
			return fmt.Errorf("writing piece %d: %w: %w", i, ErrLocal, err)
		}

		downloadStatus.CompletePiece(i)
		slog.Info("piece verified and written",
			"piece", i,
			"have", downloadStatus.DoneCount(),
			"of", downloadStatus.TotalPieces(),
			"bytes", len(piece),
			"elapsed", time.Since(started).Round(time.Second))

		// notify the peer we are interested after writing the piece
		if err = conn.Send(peer.Message{ID: peer.MsgHave, Payload: buildHavePayload(i)}); err != nil {
			return err
		}
	}

	slog.Info("download complete", "file", torrent.Name, "elapsed", time.Since(started).Round(time.Second))

	return nil
}

func buildHavePayload(pieceIndex int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(pieceIndex))
	return buf
}

// nextWantedPiece returns the lowest indexed piece we still need that the peer
// has actually advertised. Requesting anything else gets the connection closed.
func nextWantedPiece(conn *peer.PeerConnection, Done []bool) (int, bool) {
	for i := range Done {
		if !Done[i] && conn.HasPiece(i) {
			return i, true
		}
	}

	return 0, false
}

// awaitAdvertisement blocks for one message, which is how a bitfield or have
// reaches HandleMessage and widens what the peer is willing to serve.
func awaitAdvertisement(conn *peer.PeerConnection) error {
	conn.Conn.SetDeadline(time.Now().Add(120 * time.Second))

	msg, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	_, err = conn.HandleMessage(msg)

	return err
}

func downloadPiece(
	ctx context.Context,
	pieceIndex int,
	torrent types.TorrentFile,
	conn *peer.PeerConnection,
) ([]byte, error) {
	pieceLength := GetPieceLength(pieceIndex, torrent)

	piece, err := DownloadBlocks(
		ctx,
		conn,
		pieceIndex,
		pieceLength,
	)
	if err != nil {
		return nil, err
	}

	if err := VerifyPiece(piece, torrent.PieceHashes[pieceIndex]); err != nil {
		return nil, fmt.Errorf("piece %d: %w", pieceIndex, err)
	}

	return piece, nil
}
