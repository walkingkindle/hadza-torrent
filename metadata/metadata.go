// Package metadata uses the magnet link to fetch more information about the torrent file
package metadata

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"torrent-client-go/announcer"
	bencodeparser "torrent-client-go/bencode-decoder"
	"torrent-client-go/magnet-parser"
	"torrent-client-go/peer"
	"torrent-client-go/types"
)

const metadataWorkers = 20

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
						magnet,
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

func fetchMetadataFromPeer(ctx context.Context, p peer.Peer, magnet parser.MagnetURI, peerID [20]byte, infohash [20]byte) (types.TorrentFile, error) {
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

	data, ok := extension.(map[string]any)

	if !ok {
		return types.TorrentFile{}, errors.New("metadataID is not a dictionary")
	}

	for key, value := range data {
		fmt.Printf("Key: %s, Value: %v\n", key, value)
	}

	conn.Conn.Close()

	return types.TorrentFile{}, errors.New("stop here")
}

func readExtensionHandshake(ctx context.Context, conn peer.PeerConnection) (any, error) {
	msg, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	if msg.ID != peer.MsgExtended {
		return nil, fmt.Errorf("expected extended message, got %d", msg.ID)
	}

	if len(msg.Payload) == 0 {
		return nil, errors.New("empty extended message")
	}

	extensionID := msg.Payload[0]

	if extensionID != 0 {
		return nil, fmt.Errorf("expected extension handshake (ID 0), got %d", extensionID)
	}

	return bencodeparser.Decode(msg.Payload[1:])
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

func buildExtensionMessage() (any, error) {
	return bencodeparser.Encode(map[string]any{
		"m": map[string]any{
			"ut_metadata": 1,
		},
	})
}
