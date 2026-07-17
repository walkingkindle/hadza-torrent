package tracker

import (
	"encoding/binary"
	"errors"
	"net"

	"torrent-client-go/peer"
)

func parseBodyIntoPeerStruct(result any) ([]peer.Peer, error) {
	mapVal, ok := result.(map[string]any)

	if !ok {
		return []peer.Peer{}, errors.New("response not right, not a peer response")
	}

	if mapVal["peers"] == nil {
		return []peer.Peer{}, errors.New("unsupported peer response")
	}

	peerBytesStr, ok := mapVal["peers"].(string)

	if !ok {
		return []peer.Peer{}, errors.New("response not peer response")
	}

	return mapPeerBytesToPeer([]byte(peerBytesStr))
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
	}

	return peers, nil
}
