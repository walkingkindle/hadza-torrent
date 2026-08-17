// Package download handles the download feed and writes to file
package download

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"torrent-client-go/helpers"
	"torrent-client-go/peer"
	"torrent-client-go/types"
)

const stdPieceSize = 16384 // 16kib
var ErrLocal = errors.New("local failure")

func Download(conn *peer.PeerConnection, torrent types.TorrentFile, file *os.File, done []bool) error {
	// Connect leaves a short handshake deadline on the socket; being unchoked
	// can take considerably longer than that.
	conn.Conn.SetDeadline(time.Now().Add(60 * time.Second))

	err := conn.WaitForUnchoke()
	if err != nil {
		return err
	}

	defer conn.Conn.Close()

	slog.Info("starting download",
		"file", torrent.Name,
		"bytes", torrent.Length,
		"pieces", len(torrent.PieceHashes),
		"piece_length", torrent.PieceLength)

	started := time.Now()

	for helpers.CountTrue(done) < len(done) {
		i, ok := nextWantedPiece(conn, done)
		if !ok {
			slog.Info("waiting for the peer to offer a piece we still need",
				"peer", conn.Peer.IP, "advertised", conn.PieceCount(), "have", helpers.CountTrue(done))

			// Nothing on offer that we still need. A peer's availability grows
			// over the life of the connection -- a superseeding seed sends no
			// bitfield at all and reveals one piece at a time -- so wait for
			// the next advertisement instead of giving up on the whole file.
			if err := awaitAdvertisement(conn); err != nil {
				return fmt.Errorf("got %d of %d pieces from %s before it stopped offering any: %w",
					helpers.CountTrue(done), len(done), conn.Peer.IP, err)
			}
			continue
		}

		piece, err := downloadPiece(i, torrent, conn)
		if err != nil {
			return err
		}

		if _, err := file.WriteAt(piece, int64(i*torrent.PieceLength)); err != nil {
			return fmt.Errorf("writing piece %d: %w: %w", i, ErrLocal, err)
		}

		done[i] = true

		slog.Info("piece verified and written",
			"piece", i,
			"have", helpers.CountTrue(done),
			"of", len(done),
			"bytes", len(piece),
			"elapsed", time.Since(started).Round(time.Second))

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
func nextWantedPiece(conn *peer.PeerConnection, done []bool) (int, bool) {
	for i := range done {
		if !done[i] && conn.HasPiece(i) {
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

func downloadPiece(i int, torrent types.TorrentFile, conn *peer.PeerConnection) ([]byte, error) {
	pieceLength := getPieceLength(i, torrent)
	buf := make([]byte, pieceLength)

	filled := 0

	slog.Debug("downloading piece", "piece", i, "length", pieceLength, "blocks", (pieceLength+stdPieceSize-1)/stdPieceSize)

	for filled < pieceLength {
		for conn.Choked {
			slog.Debug("choked mid piece, waiting for unchoke", "piece", i, "filled", filled)

			msg, err := conn.ReadMessage()
			if err != nil {
				return nil, err
			}
			_, err = conn.HandleMessage(msg)
			if err != nil {
				return nil, err
			}
		}
		conn.Conn.SetDeadline(time.Now().Add(20 * time.Second))
		err := sendRequest(i, filled, pieceLength, conn)
		if err != nil {
			return nil, err
		}

		for {
			conn.Conn.SetDeadline(time.Now().Add(20 * time.Second))
			msg, err := conn.ReadMessage()
			if err != nil {
				return nil, err
			}

			block, err := conn.HandleMessage(msg)
			if err != nil {
				return nil, err
			}

			// A choke cancels the outstanding request, so stop waiting for a
			// block that is never going to arrive and go back to waiting for
			// an unchoke, which re-sends the request from the same offset.
			if conn.Choked {
				break
			}

			if block == nil {
				continue
			}

			if block.Index != uint32(i) {
				return nil, fmt.Errorf("peer sent piece %d, expected %d", block.Index, i)
			}

			if int(block.Begin)+len(block.Data) > pieceLength {
				return nil, fmt.Errorf("block at offset %d overruns piece %d", block.Begin, i)
			}

			copy(buf[block.Begin:], block.Data)
			filled += len(block.Data)

			slog.Debug("block received",
				"piece", i, "begin", block.Begin, "bytes", len(block.Data),
				"filled", filled, "of", pieceLength)

			break
		}
	}

	hash := sha1.Sum(buf)

	if hash != torrent.PieceHashes[i] {
		slog.Warn("piece hash mismatch",
			"piece", i,
			"got", fmt.Sprintf("%x", hash[:8]),
			"want", fmt.Sprintf("%x", torrent.PieceHashes[i][:8]))

		return nil, errors.New("piece hash mismatch")
	}

	return buf, nil
}

func getPieceLength(i int, torrent types.TorrentFile) int {
	pieceStart := i * torrent.PieceLength

	pieceEnd := pieceStart + torrent.PieceLength

	if pieceEnd > torrent.Length {
		pieceEnd = torrent.Length
	}
	thisPieceLength := pieceEnd - pieceStart

	return thisPieceLength
}

func sendRequest(i int, filled int, pieceLength int, conn *peer.PeerConnection) error {
	blockLength := min(stdPieceSize, pieceLength-filled)

	slog.Debug("requesting block", "piece", i, "begin", filled, "length", blockLength)

	return conn.Send(peer.Message{ID: peer.MsgRequest, Payload: buildPayload(i, filled, pieceLength)})
}

func buildPayload(i int, filled int, pieceLength int) []byte {
	pload := make([]byte, 12)

	binary.BigEndian.PutUint32(pload[0:4], uint32(i))

	binary.BigEndian.PutUint32(pload[4:8], uint32(filled))

	binary.BigEndian.PutUint32(pload[8:12], uint32(min(stdPieceSize, pieceLength-filled)))

	return pload
}
