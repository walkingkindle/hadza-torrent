package bencodeparser_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	bencodeparser "torrent-client-go/bencode-decoder"
)

func TestParseDict(t *testing.T) {
	tests := []struct {
		name          string
		bytes         []byte
		startPosition int
		want          map[string]any
		want2         int
		wantErr       bool
	}{
		{
			name:          "empty dict",
			bytes:         []byte("de"),
			startPosition: 0,
			want:          map[string]any{},
			want2:         2,
		},
		{
			name:          "single string value",
			bytes:         []byte("d3:cow3:mooe"),
			startPosition: 0,
			want:          map[string]any{"cow": "moo"},
			want2:         12,
		},
		{
			name:          "multiple keys mixed types",
			bytes:         []byte("d3:cow3:moo4:spami42ee"),
			startPosition: 0,
			want:          map[string]any{"cow": "moo", "spam": 42},
			want2:         22,
		},
		{
			name:          "nested list value",
			bytes:         []byte("d4:listli1ei2eee"),
			startPosition: 0,
			want:          map[string]any{"list": []any{1, 2}},
			want2:         16,
		},
		{
			name:          "nested dict value",
			bytes:         []byte("d3:keyd3:foo3:baree"),
			startPosition: 0,
			want:          map[string]any{"key": map[string]any{"foo": "bar"}},
			want2:         19,
		},
		{
			name:          "Error: does not start with 'd'",
			bytes:         []byte("i42e"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: position out of bounds",
			bytes:         []byte("de"),
			startPosition: 99,
			wantErr:       true,
		},
		{
			name:          "Error: negative position",
			bytes:         []byte("de"),
			startPosition: -1,
			wantErr:       true,
		},
		{
			name:          "Error: non-string key",
			bytes:         []byte("di1e3:fooe"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: missing value for key",
			bytes:         []byte("d3:cowe"),
			startPosition: 0,
			wantErr:       true,
		},
		{
			name:          "Error: unterminated dict",
			bytes:         []byte("d3:cow3:moo"),
			startPosition: 0,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &bencodeparser.Torrent{Data: tt.bytes}
			got, got2, gotErr := tr.ParseDict(tt.startPosition)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ParseDict() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ParseDict() succeeded unexpectedly")
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseDict() value mismatch (-want +got):\n%s", diff)
			}
			if got2 != tt.want2 {
				t.Errorf("ParseDict() position = %d, want %d", got2, tt.want2)
			}
		})
	}
}

// TestParseDictInfoSpan checks that the byte span of the "info" value is
// recorded on the Torrent, since that span feeds the info-hash in Decode.
func TestParseDictInfoSpan(t *testing.T) {
	data := []byte("d4:infod6:lengthi10eee")
	tr := &bencodeparser.Torrent{Data: data, InfoStart: -1, InfoEnd: -1}

	if _, _, err := tr.ParseDict(0); err != nil {
		t.Fatalf("ParseDict() failed: %v", err)
	}

	// The "info" value is the inner dict "d6:lengthi10ee".
	wantStart := 7
	wantEnd := 21
	if tr.InfoStart != wantStart || tr.InfoEnd != wantEnd {
		t.Errorf("info span = [%d:%d], want [%d:%d]", tr.InfoStart, tr.InfoEnd, wantStart, wantEnd)
	}
	if got := string(data[tr.InfoStart:tr.InfoEnd]); got != "d6:lengthi10ee" {
		t.Errorf("info span bytes = %q, want %q", got, "d6:lengthi10ee")
	}
}
