// Package metadata uses the magnet link to fetch more information about the torrent file
package metadata

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"torrent-client-go/announcer"
	bencodeparser "torrent-client-go/bencode-decoder"
	parser "torrent-client-go/magnet-parser"
	"torrent-client-go/peer"
	torrentparser "torrent-client-go/torrent"
	"torrent-client-go/types"
)

const (
	metadataWorkers           = 20
	metadataPieceSize         = 16 * 1024
	localMetadataID      byte = 7
	extensionHandshakeID      = 0
)

type metadataHeader struct {
	MsgType   int
	Piece     int
	TotalSize int
}

func Fetch(ctx context.Context, magnet parser.MagnetURI, peerID [20]byte) (types.TorrentFile, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	infohash, err := getInfoHashFromMagnetURI(magnet)
	if err != nil {
		return types.TorrentFile{}, err
	}
	announceResponse := announcer.AnnounceMagnet(
		ctx,
		magnet,
		string(peerID[:]),
		string(infohash[:]),
		nil,
	)
	jobs := make(chan peer.Peer)
	results := make(chan types.TorrentFile, 1)

	var wg sync.WaitGroup
	for range metadataWorkers {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return

				case p, ok := <-jobs:
					if !ok {
						return
					}

					slog.Info(
						"worker trying peer",
						"IP", p.IP,
						"port", p.Port,
					)

					torrent, err := fetchMetadataFromPeer(
						ctx,
						p,
						peerID,
						infohash,
					)
					if err != nil {
						slog.Warn(
							"metadata fetch failed",
							"IP", p.IP,
							"port", p.Port,
							"error", err,
						)
						continue
					}

					// We found the metadata!
					select {
					case results <- torrent:
						cancel()
					case <-ctx.Done():
					}

					return
				}
			}
		})
	}

	// Feed peers into workers
	go func() {
		defer close(jobs)

		for response := range announceResponse {
			for _, p := range response {
				select {
				case jobs <- p:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Close results once all workers have stopped
	go func() {
		wg.Wait()
		close(results)
	}()

	// Wait for first successful worker
	for torrent := range results {
		return torrent, nil
	}

	return types.TorrentFile{}, errors.New(
		"could not find a valid response from any of the peers",
	)
}

func fetchMetadataFromPeer(ctx context.Context, p peer.Peer, peerID [20]byte, infohash [20]byte) (types.TorrentFile, error) {
	slog.Info("received peer", "IP", p.IP, "port", p.Port)

	conn, err := peer.Connect(p, infohash, peerID)
	if err != nil {
		return types.TorrentFile{}, err
	}

	defer conn.Conn.Close()

	if !conn.SupportsExtension {
		return types.TorrentFile{}, errors.New("peer does not support extension protocol")
	}

	err = conn.SendInterested()
	if err != nil {
		return types.TorrentFile{}, err
	}

	return fetchFromPeer(ctx, conn)
}

func getInfoHashFromMagnetURI(magnet parser.MagnetURI) ([20]byte, error) {
	infohashBytes, err := hex.DecodeString(magnet.Infohash)
	if err != nil {
		return [20]byte{}, err
	}
	if len(infohashBytes) != 20 {
		return [20]byte{}, errors.New("invalid infohash length")
	}
	var infohash [20]byte
	copy(infohash[:], infohashBytes)

	return infohash, nil
}

func fetchFromPeer(ctx context.Context, conn peer.PeerConnection) (types.TorrentFile, error) {
	if err := sendExtensionHandshake(conn); err != nil {
		return types.TorrentFile{}, err
	}

	extension, err := readExtensionHandshake(ctx, conn)
	if err != nil {
		return types.TorrentFile{}, err
	}

	metadataID, metadataSize, err := getMetadataInfo(extension)
	if err != nil {
		return types.TorrentFile{}, err
	}

	slog.Info(
		"peer supports metadata",
		"metadata_id", metadataID,
		"metadata_size", metadataSize,
	)
	metadata, err := fetchMetadata(
		ctx,
		conn,
		metadataID,
		metadataSize,
	)
	if err != nil {
		return types.TorrentFile{}, err
	}

	actualHash := sha1.Sum(metadata)

	if actualHash != conn.InfoHash {
		return types.TorrentFile{}, errors.New("metadata infohash mismatch")
	}

	info, err := bencodeparser.Decode(metadata)
	if err != nil {
		return types.TorrentFile{}, fmt.Errorf(
			"failed to decode metadata: %w",
			err,
		)
	}

	infodict, ok := info.(map[string]any)
	if !ok {
		return types.TorrentFile{}, errors.New("could not parse the info metadata into dict")
	}
	return torrentparser.ParseInfoMetadata(infodict, conn.InfoHash)
}

func writeMetadataPiece(
	metadata []byte,
	pieceData []byte,
	piece int,
	metadataSize int,
) error {
	offset := piece * metadataPieceSize

	if offset >= metadataSize {
		return fmt.Errorf("piece %d is outside metadata", piece)
	}

	remaining := metadataSize - offset
	expectedSize := min(metadataPieceSize, remaining)

	if len(pieceData) != expectedSize {
		return fmt.Errorf(
			"invalid metadata piece %d size: got %d, expected %d",
			piece,
			len(pieceData),
			expectedSize,
		)
	}

	copy(metadata[offset:offset+len(pieceData)], pieceData)

	return nil
}

func fetchMetadata(
	ctx context.Context,
	conn peer.PeerConnection,
	metadataID byte,
	metadataSize int,
) ([]byte, error) {
	numPieces := (metadataSize + metadataPieceSize - 1) / metadataPieceSize

	metadata := make([]byte, metadataSize)

	for piece := range numPieces {
		data, err := fetchMetadataPiece(
			ctx,
			conn,
			metadataID,
			piece,
		)
		if err != nil {
			return nil, fmt.Errorf("fetch metadata piece %d: %w", piece, err)
		}

		if err := writeMetadataPiece(
			metadata,
			data,
			piece,
			metadataSize,
		); err != nil {
			return nil, err
		}
	}

	return metadata, nil
}

func sendMetadataRequest(
	conn peer.PeerConnection,
	peerMetadataID byte,
	piece int,
) error {
	payload, err := bencodeparser.Encode(map[string]any{
		"msg_type": 0,
		"piece":    piece,
	})
	if err != nil {
		return err
	}

	extendedPayload := make([]byte, 1+len(payload))

	extendedPayload[0] = peerMetadataID

	copy(extendedPayload[1:], payload)

	return conn.Send(peer.Message{
		ID:      peer.MsgExtended,
		Payload: extendedPayload,
	})
}

func readExtensionHandshake(
	ctx context.Context,
	conn peer.PeerConnection,
) (any, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		msg, err := conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		if msg.ID != peer.MsgExtended {
			continue
		}

		if len(msg.Payload) == 0 {
			continue
		}

		// Extension ID 0 = extension handshake.
		if msg.Payload[0] != extensionHandshakeID {
			continue
		}

		return bencodeparser.Decode(msg.Payload[1:])
	}
}

func sendExtensionHandshake(conn peer.PeerConnection) error {
	payload, err := buildExtensionMessage()
	if err != nil {
		return err
	}
	payloadByte, ok := payload.([]byte)

	if !ok {
		return errors.New("unsupported message received from bencode")
	}
	extendedPayload := make([]byte, 1+len(payloadByte))
	extendedPayload[0] = 0

	copy(extendedPayload[1:], payloadByte)

	return conn.Send(peer.Message{
		ID:      peer.MsgExtended,
		Payload: extendedPayload,
	})
}

func fetchMetadataPiece(
	ctx context.Context,
	conn peer.PeerConnection,
	peerMetadataID byte,
	piece int,
) ([]byte, error) {
	if err := sendMetadataRequest(conn, peerMetadataID, piece); err != nil {
		return nil, err
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		msg, err := conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		if msg.ID != peer.MsgExtended {
			continue
		}

		if len(msg.Payload) == 0 {
			continue
		}

		// The peer uses OUR advertised ID when sending
		// ut_metadata messages back to us.
		if msg.Payload[0] != localMetadataID {
			continue
		}

		return parseMetadataPiece(
			msg.Payload[1:],
			piece,
		)
	}
}

func parseMetadataPiece(
	payload []byte,
	expectedPiece int,
) ([]byte, error) {
	// We need to decode the bencoded header AND know
	// how many bytes were consumed by the header.
	header, consumed, err := decodeMetadataHeader(payload)
	if err != nil {
		return nil, err
	}

	if header.MsgType != 1 {
		return nil, fmt.Errorf(
			"expected metadata data message, got type %d",
			header.MsgType,
		)
	}

	if header.Piece != expectedPiece {
		return nil, fmt.Errorf(
			"expected metadata piece %d, got %d",
			expectedPiece,
			header.Piece,
		)
	}

	data := payload[consumed:]

	return data, nil
}

func getMetadataInfo(data any) (byte, int, error) {
	dict, ok := data.(map[string]any)
	if !ok {
		return 0, 0, errors.New("extension handshake is not a dictionary")
	}

	m, ok := dict["m"].(map[string]any)
	if !ok {
		return 0, 0, errors.New("extension handshake missing m dictionary")
	}

	metadataIDValue, ok := m["ut_metadata"]
	if !ok {
		return 0, 0, errors.New("peer does not support ut_metadata")
	}

	peerMetadataID, ok := metadataIDValue.(int)
	if !ok {
		return 0, 0, errors.New("invalid ut_metadata ID")
	}

	metadataSizeValue, ok := dict["metadata_size"]
	if !ok {
		return 0, 0, errors.New("peer did not provide metadata_size")
	}

	metadataSize, ok := metadataSizeValue.(int)
	if !ok {
		return 0, 0, errors.New("invalid metadata_size")
	}

	return byte(peerMetadataID), metadataSize, nil
}

func decodeMetadataHeader(data []byte) (metadataHeader, int, error) {
	value, consumed, err := bencodeparser.DecodeWithOffset(data)
	if err != nil {
		return metadataHeader{}, 0, err
	}

	dict, ok := value.(map[string]any)
	if !ok {
		return metadataHeader{}, 0, errors.New("metadata header is not a dictionary")
	}

	header := metadataHeader{}

	if value, ok := dict["msg_type"].(int); ok {
		header.MsgType = value
	} else {
		return metadataHeader{}, 0, errors.New("missing msg_type")
	}

	if value, ok := dict["piece"].(int); ok {
		header.Piece = value
	} else {
		return metadataHeader{}, 0, errors.New("missing piece")
	}

	if value, ok := dict["total_size"].(int); ok {
		header.TotalSize = value
	}

	return header, consumed, nil
}

func buildExtensionMessage() (any, error) {
	return bencodeparser.Encode(map[string]any{
		"m": map[string]any{
			"ut_metadata": int(localMetadataID),
		},
	})
}
