package bencodeparser_test

import (
	"crypto/sha1"
	"os"
	"path/filepath"
	"testing"

	bencodeparser "torrent-client-go/bencode-decoder"
)

func writeTorrent(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.torrent")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("failed to write temp torrent: %v", err)
	}
	return path
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		contents string // bencoded file contents; empty means "use a path that does not exist"
		wantLen  int    // number of top-level dict keys expected
		wantErr  bool
	}{
		{
			name:     "minimal torrent with info dict",
			contents: "d8:announce3:foo4:infod6:lengthi10eee",
			wantLen:  2,
		},
		{
			name:     "missing path",
			contents: "",
			wantErr:  true,
		},
		{
			name:     "top level is not a dict",
			contents: "i42e",
			wantErr:  true,
		},
		{
			name:     "dict without an info key",
			contents: "d8:announce3:fooe",
			wantErr:  true,
		},
		{
			name:     "malformed bencode",
			contents: "d8:announce",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location := filepath.Join(t.TempDir(), "does-not-exist.torrent")
			if tt.contents != "" {
				location = writeTorrent(t, tt.contents)
			}

			got, gotHash, gotErr := bencodeparser.Decode(location)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Decode() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Decode() succeeded unexpectedly")
			}
			if len(got) != tt.wantLen {
				t.Errorf("Decode() returned %d keys, want %d", len(got), tt.wantLen)
			}
			if gotHash == ([20]byte{}) {
				t.Error("Decode() returned a zero info hash")
			}
		})
	}
}

// TestDecodeInfoHash pins the info hash to the SHA-1 of the exact info-value
// byte span, guarding against off-by-one regressions in the recorded span.
func TestDecodeInfoHash(t *testing.T) {
	contents := "d8:announce3:foo4:infod6:lengthi10eee"
	location := writeTorrent(t, contents)

	_, gotHash, err := bencodeparser.Decode(location)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	wantHash := sha1.Sum([]byte("d6:lengthi10ee"))
	if gotHash != wantHash {
		t.Errorf("info hash = %x, want %x", gotHash, wantHash)
	}
}
