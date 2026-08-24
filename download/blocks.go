package download

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"torrent-client-go/peer"
)

type BlockDownloader struct {
	conn         *peer.PeerConnection
	pipelineSize int
	blockSize    int64
}

type blockRequest struct {
	begin  int
	length int
}

const (
	stdPieceSize = 16384
	pipelineSize = 8
)

func NewBlockDownloader(conn *peer.PeerConnection) BlockDownloader {
	return BlockDownloader{
		conn:         conn,
		pipelineSize: pipelineSize,
		blockSize:    stdPieceSize,
	}
}

func DownloadBlocks(ctx context.Context, conn *peer.PeerConnection, pieceIndex int, pieceLength int) ([]byte, error) {
	var requests []blockRequest

	for begin := 0; begin < pieceLength; begin += stdPieceSize {
		length := min(stdPieceSize, pieceLength-begin)

		requests = append(requests, blockRequest{
			begin:  begin,
			length: length,
		})
	}
	buf := make([]byte, pieceLength)

	pending := make(map[int]blockRequest)

	nextRequest := 0
	received := 0

	slog.Debug(
		"downloading piece",
		"piece", pieceIndex,
		"length", pieceLength,
		"blocks", len(requests),
	)
	for received < len(requests) {
		// Stop if the torrent has completed/cancelled.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Fill the pipeline.
		for nextRequest < len(requests) &&
			len(pending) < pipelineSize {

			// Don't send requests while we're choked.
			if conn.Choked {
				break
			}

			req := requests[nextRequest]

			conn.Conn.SetDeadline(time.Now().Add(20 * time.Second))

			if err := sendRequest(
				pieceIndex,
				req.begin,
				req.length,
				conn,
			); err != nil {
				return nil, err
			}

			pending[req.begin] = req
			nextRequest++
		}

		// If we're choked, wait for the next message.
		if conn.Choked {
			slog.Debug(
				"choked while downloading piece",
				"piece", pieceIndex,
				"received", received,
				"pending", len(pending),
			)

			conn.Conn.SetDeadline(time.Now().Add(20 * time.Second))

			msg, err := conn.ReadMessage()
			if err != nil {
				return nil, err
			}

			if _, err := conn.HandleMessage(msg); err != nil {
				return nil, err
			}

			continue
		}

		// Read the next message from the peer.
		conn.Conn.SetDeadline(time.Now().Add(20 * time.Second))

		msg, err := conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		block, err := conn.HandleMessage(msg)
		if err != nil {
			return nil, err
		}

		// The message wasn't a PIECE message.
		if block == nil {
			continue
		}

		// Make sure the block belongs to the piece we're downloading.
		if block.Index != uint32(pieceIndex) {
			return nil, fmt.Errorf(
				"peer sent piece %d, expected %d",
				block.Index,
				pieceIndex,
			)
		}

		begin := int(block.Begin)

		// Make sure this is a block we actually requested.
		req, ok := pending[begin]
		if !ok {
			return nil, fmt.Errorf(
				"received unexpected block at offset %d",
				begin,
			)
		}

		// Make sure the peer returned the expected amount of data.
		if len(block.Data) != req.length {
			return nil, fmt.Errorf(
				"block at offset %d: got %d bytes, expected %d",
				begin,
				len(block.Data),
				req.length,
			)
		}

		copy(buf[begin:], block.Data)

		delete(pending, begin)
		received++

		slog.Debug(
			"block received",
			"piece", pieceIndex,
			"begin", begin,
			"bytes", len(block.Data),
			"received", received,
			"of", len(requests),
			"pending", len(pending),
		)
	}
	return buf, nil
}
