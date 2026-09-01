// Package metadata uses the magnet link to fetch more information about the torrent file
package metadata

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"torrent-client-go/announcer"
	bencodeparser "torrent-client-go/bencode-decoder"
	"torrent-client-go/magnet-parser"
	"torrent-client-go/peer"
	"torrent-client-go/types"
)

func Fetch(ctx context.Context, magnet parser.MagnetURI, peerID [20]byte) (types.TorrentFile, error) {
	announceResponse := announcer.AnnounceMagnet(
		ctx,
		magnet,
		string(peerID[:]),
	)
	for response := range announceResponse {
		for _, p := range response {
			var infohash [20]byte
			copy(infohash[:], []byte(magnet.Infohash))
			conn, err := peer.Connect(p, infohash, peerID)
			if err != nil {
				continue
			}

			if !conn.SupportsExtension {
				slog.Warn("this peer does not support extension, continuing")
				conn.Conn.Close()
				continue
			}

			err = conn.SendInterested()
			if err != nil {
				slog.Warn("peer connection suceeded but not able to send interested")
				conn.Conn.Close()
				continue
			}

			torrent, err := fetchFromPeer(ctx, conn)
			conn.Conn.Close()

			if err != nil {
				slog.Warn("failed to fetch metadata", "peer", p.IP, "error", err)
				continue
			}

			return torrent, nil
		}
	}
	return types.TorrentFile{}, errors.New("could not find a valid response from any of the peers")
}

func fetchFromPeer(ctx context.Context, conn peer.PeerConnection) (types.TorrentFile, error) {
	if err := sendExtensionHandshake(conn); err != nil {
		return types.TorrentFile{}, err
	}

	extension, err := readExtensionHandshake(ctx, conn)
	if err != nil {
		return types.TorrentFile{}, err
	}

	_, ok := extension.(map[string]any)

	if !ok {
		return types.TorrentFile{}, errors.New("metadataID is not a dictionary")
	}

	// for key, value := range data {
	// 	fmt.Printf("Key: %s, Value: %v\n", key, value)
	// }

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
