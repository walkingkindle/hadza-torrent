package peer_test

import (
	"testing"

	"torrent-client-go/peer"
)

// buildHandshake returns a well-formed 68-byte handshake with the given
// info hash and peer ID.
func buildHandshake(infoHash, peerID [20]byte) []byte {
	data := make([]byte, 68)
	data[0] = 19
	copy(data[1:20], "BitTorrent protocol")
	// data[20:28] reserved, left zero
	copy(data[28:48], infoHash[:])
	copy(data[48:68], peerID[:])
	return data
}

func TestParseHandshake(t *testing.T) {
	var infoHash, peerID [20]byte
	for i := 0; i < 20; i++ {
		infoHash[i] = byte(i)
		peerID[i] = byte(100 + i)
	}

	data := buildHandshake(infoHash, peerID)

	h, err := peer.ParseHandshake(data)
	if err != nil {
		t.Fatalf("ParseHandshake() returned error: %v", err)
	}

	if h.InfoHash != infoHash {
		t.Errorf("InfoHash = %v, want %v", h.InfoHash, infoHash)
	}
	if h.PeerID != peerID {
		t.Errorf("PeerID = %v, want %v", h.PeerID, peerID)
	}
}

// Serialize followed by ParseHandshake must recover the original fields.
func TestParseHandshakeRoundTrip(t *testing.T) {
	var infoHash, peerID [20]byte
	for i := 0; i < 20; i++ {
		infoHash[i] = byte(200 - i)
		peerID[i] = byte(50 + i)
	}

	original := peer.Handshake{InfoHash: infoHash, PeerID: peerID}

	serialized, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() returned error: %v", err)
	}

	got, err := peer.ParseHandshake(serialized[:])
	if err != nil {
		t.Fatalf("ParseHandshake() returned error: %v", err)
	}

	if got != original {
		t.Errorf("round trip = %+v, want %+v", got, original)
	}
}

func TestParseHandshakeErrors(t *testing.T) {
	valid := buildHandshake([20]byte{}, [20]byte{})

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "nil data",
			data: nil,
		},
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "too short",
			data: valid[:67],
		},
		{
			name: "wrong pstrlen",
			data: func() []byte {
				d := buildHandshake([20]byte{}, [20]byte{})
				d[0] = 18
				return d
			}(),
		},
		{
			name: "wrong protocol string",
			data: func() []byte {
				d := buildHandshake([20]byte{}, [20]byte{})
				copy(d[1:20], "NotTorrent protocol")
				return d
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := peer.ParseHandshake(tt.data); err == nil {
				t.Errorf("ParseHandshake(%s) = nil error, want error", tt.name)
			}
		})
	}
}

// Data longer than 68 bytes (e.g. a handshake with trailing message bytes)
// should still parse, reading only the fixed handshake fields.
func TestParseHandshakeExtraBytes(t *testing.T) {
	var infoHash, peerID [20]byte
	for i := 0; i < 20; i++ {
		infoHash[i] = byte(i)
		peerID[i] = byte(i)
	}

	data := buildHandshake(infoHash, peerID)
	data = append(data, 0xde, 0xad, 0xbe, 0xef)

	h, err := peer.ParseHandshake(data)
	if err != nil {
		t.Fatalf("ParseHandshake() returned error: %v", err)
	}

	if h.InfoHash != infoHash {
		t.Errorf("InfoHash = %v, want %v", h.InfoHash, infoHash)
	}
	if h.PeerID != peerID {
		t.Errorf("PeerID = %v, want %v", h.PeerID, peerID)
	}
}
