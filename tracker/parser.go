package tracker

import (
	"encoding/binary"
	"errors"
	"log/slog"
	"net"

	"torrent-client-go/peer"
)

func parseBodyIntoPeerStruct(result any) (*peer.TrackersResponse, error) {
	mapVal, ok := result.(map[string]any)

	if !ok {
		return nil, errors.New("response not right, not a peer response")
	}

	if reason, failed := mapVal["failure reason"]; failed {
		return nil, errors.New("tracker refused the announce: " + toString(reason))
	}

	// The swarm counts are worth surfacing: they are how you tell "nobody is
	// sharing this" apart from "the tracker is only handing me a few peers".
	slog.Info("tracker announce",
		"seeders", mapVal["complete"],
		"leechers", mapVal["incomplete"],
		"reannounce_interval_s", mapVal["interval"])

	if mapVal["peers"] == nil || mapVal["interval"] == nil {
		return nil, errors.New("unsupported peer response")
	}

	peerString, interval, err := assertTrackerResponse(mapVal)
	if err != nil {
		return nil, err
	}

	slog.Info("tracker peer payload", "bytes", len(peerString))

	peers, err := mapPeerBytesToPeer([]byte(peerString))
	if err != nil {
		return nil, err
	}
	return &peer.TrackersResponse{Interval: interval, Peers: peers}, nil
}

func assertTrackerResponse(mapVal map[string]any) (string, int, error) {
	peerBytesStr, ok := mapVal["peers"].(string)
	if !ok {
		return "", 0, errors.New("Error while formatting peers response")
	}
	interval, ok1 := mapVal["interval"].(int)

	if !ok1 {
		return "", 0, errors.New("Error while formatting peers response")
	}

	return peerBytesStr, interval, nil
}

func mapPeerBytesToPeer(peerBytes []byte) ([]peer.Peer, error) {
	const peerSize = 6

	numPeers := len(peerBytes) / peerSize

	if len(peerBytes)%peerSize != 0 {
		return []peer.Peer{}, errors.New("received malformed peers")
	}

	peers := make([]peer.Peer, numPeers)

	for i := range numPeers {
		offset := i * peerSize

		peers[i].IP = net.IPv4(peerBytes[offset], peerBytes[offset+1], peerBytes[offset+2], peerBytes[offset+3])
		peers[i].Port = binary.BigEndian.Uint16(peerBytes[offset+4 : offset+6])

		slog.Debug("tracker peer", "index", i, "addr", peers[i].IP.String(), "port", peers[i].Port)
	}

	slog.Info("tracker returned peers", "count", numPeers)

	return peers, nil
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}

	return "unknown reason"
}
