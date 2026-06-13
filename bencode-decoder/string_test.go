package bencodeparser_test

import (
	"testing"

	"torrent-client-go/bencode-decoder"
)

func TestParseString(t *testing.T) {
	tests := []struct {
		name     string
		position int
		bytes    []byte
		want     string
		want2    int
		wantErr  bool
	}{
		// --- happy path ---
		{
			name:     "simple string",
			position: 0,
			bytes:    []byte("8:announce"),
			want:     "announce",
			want2:    10, // cursor lands just past the last content byte
		},
		{
			name:     "single char",
			position: 0,
			bytes:    []byte("1:a"),
			want:     "a",
			want2:    3,
		},
		{
			name:     "empty string is valid",
			position: 0,
			bytes:    []byte("0:"),
			want:     "",
			want2:    2,
		},
		{
			name:     "multi-digit length",
			position: 0,
			bytes:    []byte("11:hello world"),
			want:     "hello world",
			want2:    14,
		},
		// --- the value is parsed by the SAME function as the key ---
		{
			name:     "url-like value with colons inside",
			position: 0,
			bytes:    []byte("36:udp://tracker.opentrackr.org:1337/ann"[:39]),
			want:     "udp://tracker.opentrackr.org:1337/an",
			want2:    39,
		},
		{
			name:     "content containing structural bytes",
			position: 0,
			bytes:    []byte("5:di:ee"), // the 5 content bytes are 'd','i',':','e','e'
			want:     "di:ee",
			want2:    7,
		},
		// --- error cases ---
		{
			name:     "does not start with digit",
			position: 0,
			bytes:    []byte("i42e"),
			wantErr:  true,
		},
		{
			name:     "position out of bounds",
			position: 99,
			bytes:    []byte("8:announce"),
			wantErr:  true,
		},
		{
			name:     "negative position",
			position: -1,
			bytes:    []byte("8:announce"),
			wantErr:  true,
		},
		{
			name:     "no colon",
			position: 0,
			bytes:    []byte("8announce"),
			wantErr:  true,
		},
		{
			name:     "non-numeric length",
			position: 0,
			bytes:    []byte("x:foo"),
			wantErr:  true, // caught by the digit guard, actually
		},
		{
			name:     "length exceeds available bytes",
			position: 0,
			bytes:    []byte("99:short"),
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &bencodeparser.Torrent{Data: tt.bytes}
			got, got2, gotErr := tr.ParseString(tt.position)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ParseString() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ParseString() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("ParseString() value = %q, want %q", got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("ParseString() position = %d, want %d", got2, tt.want2)
			}
		})
	}
}
