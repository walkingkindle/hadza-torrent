package bencodeparser_test

import (
	"testing"

	bencodeparser "torrent-client-go/bencode-decoder"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		name          string
		bytesArr      []byte
		startPosition int
		want          int
		want2         int
		wantErr       bool
	}{
		{
			name:          "Valid positive integer",
			bytesArr:      []byte("i42e"),
			startPosition: 0,
			want:          42,
			want2:         4,
			wantErr:       false,
		},
		{
			name:          "Valid negative integer",
			bytesArr:      []byte("i-42e"),
			startPosition: 0,
			want:          -42,
			want2:         5,
			wantErr:       false,
		},
		{
			name:          "Valid zero",
			bytesArr:      []byte("i0e"),
			startPosition: 0,
			want:          0,
			want2:         3,
			wantErr:       false,
		},
		{
			name:          "Parses at a non-zero start position",
			bytesArr:      []byte("li42ee"),
			startPosition: 1,
			want:          42,
			want2:         5,
			wantErr:       false,
		},
		{
			name:          "Error: Missing leading 'i'",
			bytesArr:      []byte("42e"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: Missing trailing 'e'",
			bytesArr:      []byte("i42"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: Invalid negative zero",
			bytesArr:      []byte("i-0e"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: Leading zeroes",
			bytesArr:      []byte("i03e"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: Empty integer body",
			bytesArr:      []byte("ie"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: Non-numeric data",
			bytesArr:      []byte("i42abc2e"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: lone minus sign",
			bytesArr:      []byte("i-e"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: position out of bounds",
			bytesArr:      []byte("i42e"),
			startPosition: 99,
			wantErr:       true,
		},
		{
			name:          "Error: negative position",
			bytesArr:      []byte("i42e"),
			startPosition: -1,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &bencodeparser.Torrent{Data: tt.bytesArr}
			got, got2, gotErr := tr.ParseInt(tt.startPosition)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ParseInt() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ParseInt() succeeded unexpectedly")
			}

			if got != tt.want {
				t.Errorf("ParseInt() value got = %v, want %v", got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("ParseInt() next index got = %v, want %v", got2, tt.want2)
			}
		})
	}
}
