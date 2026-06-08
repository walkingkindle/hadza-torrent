package bencodeparser_test

import (
	"testing"

	"torrent-client-go/bencode-decoder"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		location string
		want     []byte
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := bencodeparser.Decode(tt.location)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Decode() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Decode() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}
