package peer_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"torrent-client-go/peer"
)

// buildPiecePayload returns the payload of a piece message: a 4-byte index,
// a 4-byte begin offset, followed by the block data.
func buildPiecePayload(index, begin uint32, data []byte) []byte {
	payload := make([]byte, 8+len(data))
	binary.BigEndian.PutUint32(payload[0:4], index)
	binary.BigEndian.PutUint32(payload[4:8], begin)
	copy(payload[8:], data)
	return payload
}

func TestParsePiece(t *testing.T) {
	tests := []struct {
		name  string
		index uint32
		begin uint32
		data  []byte
	}{
		{
			name:  "typical block",
			index: 3,
			begin: 16384,
			data:  []byte("hello world"),
		},
		{
			name:  "zero index and begin",
			index: 0,
			begin: 0,
			data:  []byte{0x00, 0xff, 0x10},
		},
		{
			name:  "max index and begin",
			index: 0xffffffff,
			begin: 0xffffffff,
			data:  []byte{0x01},
		},
		{
			name:  "full 16KiB block",
			index: 12,
			begin: 32768,
			data:  bytes.Repeat([]byte{0xab}, 16384),
		},
		{
			name:  "empty data",
			index: 1,
			begin: 2,
			data:  []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := peer.ParsePiece(buildPiecePayload(tt.index, tt.begin, tt.data))
			if err != nil {
				t.Fatalf("ParsePiece() returned error: %v", err)
			}

			if got.Index != tt.index {
				t.Errorf("Index = %d, want %d", got.Index, tt.index)
			}
			if got.Begin != tt.begin {
				t.Errorf("Begin = %d, want %d", got.Begin, tt.begin)
			}
			if !bytes.Equal(got.Data, tt.data) {
				t.Errorf("Data = %v, want %v", got.Data, tt.data)
			}
		})
	}
}

func TestParsePieceErrors(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "nil payload",
			payload: nil,
		},
		{
			name:    "empty payload",
			payload: []byte{},
		},
		{
			name:    "index only",
			payload: []byte{0, 0, 0, 1},
		},
		{
			name:    "one byte short of header",
			payload: []byte{0, 0, 0, 1, 0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := peer.ParsePiece(tt.payload)
			if err == nil {
				t.Fatalf("ParsePiece(%s) = nil error, want error", tt.name)
			}
			if got.Index != 0 || got.Begin != 0 || got.Data != nil {
				t.Errorf("ParsePiece(%s) = %+v, want zero PieceBlock", tt.name, got)
			}
		})
	}
}

// A payload with exactly the 8-byte header and no block data is not an error;
// it parses to an empty (but valid) block.
func TestParsePieceHeaderOnly(t *testing.T) {
	got, err := peer.ParsePiece([]byte{0, 0, 0, 7, 0, 0, 0x40, 0x00})
	if err != nil {
		t.Fatalf("ParsePiece() returned error: %v", err)
	}

	if got.Index != 7 {
		t.Errorf("Index = %d, want 7", got.Index)
	}
	if got.Begin != 16384 {
		t.Errorf("Begin = %d, want 16384", got.Begin)
	}
	if len(got.Data) != 0 {
		t.Errorf("Data = %v, want empty", got.Data)
	}
}

// Data must be read big-endian, not little-endian: a payload whose bytes
// differ under the two encodings pins down the byte order.
func TestParsePieceByteOrder(t *testing.T) {
	payload := []byte{0x00, 0x00, 0x01, 0x02, 0x00, 0x00, 0x40, 0x00, 0x99}

	got, err := peer.ParsePiece(payload)
	if err != nil {
		t.Fatalf("ParsePiece() returned error: %v", err)
	}

	if got.Index != 0x0102 {
		t.Errorf("Index = %#x, want %#x", got.Index, 0x0102)
	}
	if got.Begin != 0x4000 {
		t.Errorf("Begin = %#x, want %#x", got.Begin, 0x4000)
	}
}

// ParsePiece does not copy the block: Data aliases the payload it was given.
// ReadMessage allocates a fresh buffer per message today, so this is safe, but
// the returned block must not outlive a reused read buffer.
func TestParsePieceDataAliasesPayload(t *testing.T) {
	payload := buildPiecePayload(0, 0, []byte{1, 2, 3})

	got, err := peer.ParsePiece(payload)
	if err != nil {
		t.Fatalf("ParsePiece() returned error: %v", err)
	}

	payload[8] = 0xff
	if got.Data[0] != 0xff {
		t.Skip("ParsePiece now copies the block data; aliasing note in ReadMessage no longer applies")
	}
}
