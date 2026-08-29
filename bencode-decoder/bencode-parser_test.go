package bencodeparser_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	bencodeparser "torrent-client-go/bencode-decoder"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    any
		wantErr bool
	}{
		{
			name: "top-level dict",
			data: []byte("d8:announce3:foo4:infod6:lengthi10eee"),
			want: map[string]any{
				"announce": "foo",
				"info":     map[string]any{"length": 10},
			},
		},
		{
			name: "top-level int",
			data: []byte("i42e"),
			want: 42,
		},
		{
			name: "top-level string",
			data: []byte("4:spam"),
			want: "spam",
		},
		{
			name: "top-level list",
			data: []byte("l4:spami42ee"),
			want: []any{"spam", 42},
		},
		{
			name:    "empty input",
			data:    []byte(""),
			wantErr: true,
		},
		{
			name:    "malformed bencode",
			data:    []byte("d8:announce"),
			wantErr: true,
		},
		{
			name:    "unknown type byte",
			data:    []byte("x"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := bencodeparser.Decode(tt.data)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Decode() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Decode() succeeded unexpectedly")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Decode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name    string
		data    any
		want    []byte
		wantErr bool
	}{
		{
			name:    "success simple int",
			data:    6,
			wantErr: false,
			want:    []byte("i6e"),
		},
		{
			name:    "success simple string",
			data:    "spam",
			wantErr: false,
			want:    []byte("4:spam"),
		},
		{
			name: "success simple dict",
			data: map[string]any{
				"foo": 6,
			},
			wantErr: false,
			want:    []byte("d3:fooi6ee"),
		},
		{
			name: "success nested dict",
			data: map[string]any{
				"foo": map[string]any{
					"bar": 6,
				},
			},
			wantErr: false,
			want:    []byte("d3:food3:bari6eee"),
		},
		{
			name: "success extension handshake",
			data: map[string]any{
				"m": map[string]any{
					"ut_metadata": 1,
				},
			},
			wantErr: false,
			want:    []byte("d1:md11:ut_metadatai1eee"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := bencodeparser.Encode(tt.data)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Encode() failed: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("Encode() succeeded unexpectedly")
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Encode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
