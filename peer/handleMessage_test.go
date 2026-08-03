package peer_test

import (
	"bytes"
	"testing"

	"torrent-client-go/peer"
)

// A piece message must hand the parsed block back to the caller.
func TestHandleMessagePiece(t *testing.T) {
	pc := &peer.PeerConnection{}
	data := []byte("block data")

	block, err := pc.HandleMessage(&peer.Message{
		ID:      7,
		Payload: buildPiecePayload(4, 16384, data),
	})
	if err != nil {
		t.Fatalf("HandleMessage() returned error: %v", err)
	}

	if block.Index != 4 {
		t.Errorf("Index = %d, want 4", block.Index)
	}
	if block.Begin != 16384 {
		t.Errorf("Begin = %d, want 16384", block.Begin)
	}
	if !bytes.Equal(block.Data, data) {
		t.Errorf("Data = %q, want %q", block.Data, data)
	}
}

// A malformed piece payload must surface as an error, not a silent zero block.
func TestHandleMessagePieceMalformed(t *testing.T) {
	pc := &peer.PeerConnection{}

	if _, err := pc.HandleMessage(&peer.Message{ID: 7, Payload: []byte{0, 0, 0}}); err == nil {
		t.Error("HandleMessage() = nil error, want error for short piece payload")
	}
}

func TestHandleMessageState(t *testing.T) {
	pc := &peer.PeerConnection{}

	if _, err := pc.HandleMessage(&peer.Message{ID: 0}); err != nil {
		t.Fatalf("choke returned error: %v", err)
	}
	if !pc.Choked {
		t.Error("Choked = false after choke, want true")
	}

	if _, err := pc.HandleMessage(&peer.Message{ID: 1}); err != nil {
		t.Fatalf("unchoke returned error: %v", err)
	}
	if pc.Choked {
		t.Error("Choked = true after unchoke, want false")
	}

	bitfield := []byte{0xff, 0x0f}
	if _, err := pc.HandleMessage(&peer.Message{ID: 5, Payload: bitfield}); err != nil {
		t.Fatalf("bitfield returned error: %v", err)
	}
	if !bytes.Equal(pc.Bitfield, bitfield) {
		t.Errorf("Bitfield = %v, want %v", pc.Bitfield, bitfield)
	}
}
