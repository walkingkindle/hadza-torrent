package bencodeparser_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	bencodeparser "torrent-client-go/bencode-decoder"
)

func TestParseList(t *testing.T) {
	tests := []struct {
		name          string
		bytes         []byte
		startPosition int
		want          []any
		want2         int
		wantErr       bool
	}{
		{
			name:          "empty list",
			bytes:         []byte("le"),
			startPosition: 0,
			want:          []any{},
			want2:         2,
		},
		{
			name:          "list of ints",
			bytes:         []byte("li1ei2ei3ee"),
			startPosition: 0,
			want:          []any{1, 2, 3},
			want2:         11,
		},
		{
			name:          "list of strings",
			bytes:         []byte("l4:spam4:eggse"),
			startPosition: 0,
			want:          []any{"spam", "eggs"},
			want2:         14,
		},
		{
			name:          "mixed list",
			bytes:         []byte("l4:spami42ee"),
			startPosition: 0,
			want:          []any{"spam", 42},
			want2:         12,
		},
		{
			name:          "nested list",
			bytes:         []byte("lli1eee"),
			startPosition: 0,
			want:          []any{[]any{1}},
			want2:         7,
		},
		{
			name:          "list with trailing data parses only the list",
			bytes:         []byte("li1eei2e"),
			startPosition: 0,
			want:          []any{1},
			want2:         5,
		},
		{
			name:          "Error: does not start with 'l'",
			bytes:         []byte("i42e"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: position out of bounds",
			bytes:         []byte("le"),
			startPosition: 99,
			wantErr:       true,
		},
		{
			name:          "Error: negative position",
			bytes:         []byte("le"),
			startPosition: -1,
			wantErr:       true,
		},
		{
			name:          "Error: unterminated list",
			bytes:         []byte("li1e"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: malformed element",
			bytes:         []byte("li-0ee"),
			startPosition: 0,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &bencodeparser.Torrent{Data: tt.bytes}
			got, got2, gotErr := tr.ParseList(tt.startPosition)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ParseList() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ParseList() succeeded unexpectedly")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseList() value mismatch (-want +got):\n%s", diff)
			}
			if got2 != tt.want2 {
				t.Errorf("ParseList() position = %d, want %d", got2, tt.want2)
			}
		})
	}
}
