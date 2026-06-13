package bencodeparser_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	bencodeparser "torrent-client-go/bencode-decoder"
)

func TestDispatch(t *testing.T) {
	tests := []struct {
		name          string
		bytes         []byte
		startPosition int
		want          any
		want2         int
		wantErr       bool
	}{
		{
			name:          "routes to int",
			bytes:         []byte("i42e"),
			startPosition: 0,
			want:          42,
			want2:         4,
		},
		{
			name:          "routes to string",
			bytes:         []byte("4:spam"),
			startPosition: 0,
			want:          "spam",
			want2:         6,
		},
		{
			name:          "routes to list",
			bytes:         []byte("li1ee"),
			startPosition: 0,
			want:          []any{1},
			want2:         5,
		},
		{
			name:          "routes to dict",
			bytes:         []byte("d3:cow3:mooe"),
			startPosition: 0,
			want:          map[string]any{"cow": "moo"},
			want2:         12,
		},
		{
			name:          "Error: position out of bounds",
			bytes:         []byte("i42e"),
			startPosition: 99,
			wantErr:       true,
		},
		{
			name:          "Error: negative position",
			bytes:         []byte("i42e"),
			startPosition: -1,
			wantErr:       true,
		},
		{
			name:          "Error: unknown type byte",
			bytes:         []byte("xyz"),
			startPosition: 0,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &bencodeparser.Torrent{Data: tt.bytes}
			got, got2, gotErr := tr.Dispatch(tt.startPosition)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Dispatch() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Dispatch() succeeded unexpectedly")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Dispatch() value mismatch (-want +got):\n%s", diff)
			}
			if got2 != tt.want2 {
				t.Errorf("Dispatch() position = %d, want %d", got2, tt.want2)
			}
		})
	}
}
