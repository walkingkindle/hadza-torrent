package peer_test

import (
	"bytes"
	"testing"

	"torrent-client-go/peer"
)

func TestSerialize(t *testing.T) {
	infoHash := [20]byte{}
	peerID := [20]byte{}
	for i := 0; i < 20; i++ {
		infoHash[i] = byte(i)
		peerID[i] = byte(20 + i)
	}

	h := peer.Handshake{InfoHash: infoHash, PeerID: peerID}

	got, err := h.Serialize()
	if err != nil {
		t.Fatalf("Serialize() returned error: %v", err)
	}

	// Byte 0: length of the protocol string ("BitTorrent protocol" == 19).
	if got[0] != 19 {
		t.Errorf("pstrlen = %d, want 19", got[0])
	}

	// Bytes 1..20: the protocol string.
	if pstr := string(got[1:20]); pstr != "BitTorrent protocol" {
		t.Errorf("pstr = %q, want %q", pstr, "BitTorrent protocol")
	}

	// Bytes 20..28: 8 reserved bytes, all zero.
	if reserved := got[20:28]; !bytes.Equal(reserved, make([]byte, 8)) {
		t.Errorf("reserved = %v, want 8 zero bytes", reserved)
	}

	// Bytes 28..48: info hash.
	if hash := got[28:48]; !bytes.Equal(hash, infoHash[:]) {
		t.Errorf("info hash = %v, want %v", hash, infoHash)
	}

	// Bytes 48..68: peer ID.
	if id := got[48:68]; !bytes.Equal(id, peerID[:]) {
		t.Errorf("peer ID = %v, want %v", id, peerID)
	}
}

func TestSerializeLength(t *testing.T) {
	var h peer.Handshake

	got, err := h.Serialize()
	if err != nil {
		t.Fatalf("Serialize() returned error: %v", err)
	}

	if len(got) != 68 {
		t.Errorf("serialized length = %d, want 68", len(got))
	}
}

func TestSerializeZeroValueIsAllZeroPayload(t *testing.T) {
	var h peer.Handshake

	got, err := h.Serialize()
	if err != nil {
		t.Fatalf("Serialize() returned error: %v", err)
	}

	// With a zero-value handshake, everything after the protocol header
	// (info hash + peer ID + reserved) must be zero.
	if tail := got[20:]; !bytes.Equal(tail, make([]byte, len(tail))) {
		t.Errorf("payload for zero-value handshake = %v, want all zeros", tail)
	}
}
