package peer_test

import (
	"testing"

	"torrent-client-go/peer"
)

func TestHasPieceFromBitfield(t *testing.T) {
	// 0b10110000, 0b00000001 -> pieces 0, 2, 3 and 15
	pc := peer.PeerConnection{Bitfield: []byte{0xB0, 0x01}}

	want := map[int]bool{
		0: true, 1: false, 2: true, 3: true,
		4: false, 5: false, 6: false, 7: false,
		8: false, 15: true,
	}

	for i, expected := range want {
		if got := pc.HasPiece(i); got != expected {
			t.Errorf("HasPiece(%d) = %t, want %t", i, got, expected)
		}
	}
}

func TestHasPieceOutOfRange(t *testing.T) {
	pc := peer.PeerConnection{Bitfield: []byte{0xFF}}

	for _, i := range []int{-1, 8, 100} {
		if pc.HasPiece(i) {
			t.Errorf("HasPiece(%d) = true, want false for out of range index", i)
		}
	}
}

func TestHasPieceNilBitfield(t *testing.T) {
	pc := peer.PeerConnection{}

	if pc.HasPiece(0) {
		t.Error("HasPiece(0) = true on a nil bitfield, want false")
	}
}

// A peer may send have messages without ever sending a bitfield, so the
// bitfield has to grow to fit instead of panicking.
func TestHaveGrowsBitfield(t *testing.T) {
	pc := peer.PeerConnection{}

	msg := &peer.Message{ID: peer.MsgHave, Payload: []byte{0, 0, 0x4C, 0x36}} // piece 19510
	if _, err := pc.HandleMessage(msg); err != nil {
		t.Fatalf("HandleMessage(have) = %v, want nil", err)
	}

	if !pc.HasPiece(19510) {
		t.Error("HasPiece(19510) = false after a have message, want true")
	}
	if pc.HasPiece(19509) || pc.HasPiece(19511) {
		t.Error("have message set neighbouring bits")
	}
}

func TestHaveKeepsExistingBitfield(t *testing.T) {
	pc := peer.PeerConnection{Bitfield: []byte{0x80}} // piece 0

	msg := &peer.Message{ID: peer.MsgHave, Payload: []byte{0, 0, 0, 9}}
	if _, err := pc.HandleMessage(msg); err != nil {
		t.Fatalf("HandleMessage(have) = %v, want nil", err)
	}

	if !pc.HasPiece(0) {
		t.Error("growing the bitfield dropped piece 0")
	}
	if !pc.HasPiece(9) {
		t.Error("HasPiece(9) = false after a have message, want true")
	}
}

func TestHaveRejectsShortPayload(t *testing.T) {
	pc := peer.PeerConnection{}

	msg := &peer.Message{ID: peer.MsgHave, Payload: []byte{0, 0, 1}}
	if _, err := pc.HandleMessage(msg); err == nil {
		t.Error("HandleMessage(short have) = nil error, want error")
	}
}
