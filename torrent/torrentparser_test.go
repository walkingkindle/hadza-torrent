package torrentparser

import (
	"crypto/sha1"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// benStr encodes s as a bencoded string ("<len>:<s>").
func benStr(s string) string {
	return strconv.Itoa(len(s)) + ":" + s
}

// twentyA / twentyB are single 20-byte piece hashes.
var (
	twentyA = strings.Repeat("a", 20)
	twentyB = strings.Repeat("b", 20)
)

// buildInfo builds a bencoded info dict. Keys are kept in lexical order, as the
// bencode spec requires. pieces is the raw concatenation of 20-byte hashes.
func buildInfo(name string, length, pieceLength int, pieces string) string {
	return "d" +
		benStr("length") + "i" + strconv.Itoa(length) + "e" +
		benStr("name") + benStr(name) +
		benStr("piece length") + "i" + strconv.Itoa(pieceLength) + "e" +
		benStr("pieces") + benStr(pieces) +
		"e"
}

func buildTorrent(announce, info string) string {
	return "d" + benStr("announce") + benStr(announce) + benStr("info") + info + "e"
}

func TestParseTorrentFile(t *testing.T) {
	info := buildInfo("debian.iso", 262144, 32768, twentyA+twentyB)
	announce := "http://tracker.example.com:8080/announce"
	data := []byte(buildTorrent(announce, info))

	got, err := ParseTorrentFile(data)
	if err != nil {
		t.Fatalf("ParseTorrentFile() failed: %v", err)
	}

	if got.Name != "debian.iso" {
		t.Errorf("Name = %q, want %q", got.Name, "debian.iso")
	}
	if got.Announce != announce {
		t.Errorf("Announce = %q, want %q", got.Announce, announce)
	}
	if got.Length != 262144 {
		t.Errorf("Length = %d, want %d", got.Length, 262144)
	}
	if got.PieceLength != 32768 {
		t.Errorf("PieceLength = %d, want %d", got.PieceLength, 32768)
	}

	wantHashes := [][20]byte{[20]byte([]byte(twentyA)), [20]byte([]byte(twentyB))}
	if diff := cmp.Diff(wantHashes, got.PieceHashes); diff != "" {
		t.Errorf("PieceHashes mismatch (-want +got):\n%s", diff)
	}

	wantHash := sha1.Sum([]byte(info))
	if got.InfoHash != wantHash {
		t.Errorf("InfoHash = %x, want %x", got.InfoHash, wantHash)
	}
}

func TestParseTorrentFileErrors(t *testing.T) {
	validInfo := buildInfo("a", 1, 1, twentyA)

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "top level is not a dict",
			data: []byte("i42e"),
		},
		{
			name: "no info key",
			data: []byte("d8:announce3:fooe"),
		},
		{
			name: "info missing name",
			data: []byte("d8:announce3:foo4:infod6:lengthi1eee"),
		},
		{
			name: "pieces not a multiple of 20",
			data: []byte(buildTorrent("http://x", buildInfo("a", 1, 1, "short"))),
		},
		{
			name: "malformed bencode",
			data: []byte("d8:announce"),
		},
		{
			name: "truncated data",
			data: []byte(buildTorrent("http://x", validInfo)[:10]),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseTorrentFile(tt.data); err == nil {
				t.Error("ParseTorrentFile() succeeded unexpectedly")
			}
		})
	}
}

func TestFindInfoSlice(t *testing.T) {
	info := buildInfo("x", 1, 1, twentyA)
	data := []byte(buildTorrent("http://tracker", info))

	start, end, err := findInfoSlice(data)
	if err != nil {
		t.Fatalf("findInfoSlice() failed: %v", err)
	}
	if got := string(data[start:end]); got != info {
		t.Errorf("info span = %q, want %q", got, info)
	}
}

func TestFindInfoSliceErrors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "not a dict", data: []byte("i42e")},
		{name: "no info key", data: []byte("d3:cow3:mooe")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := findInfoSlice(tt.data); err == nil {
				t.Error("findInfoSlice() succeeded unexpectedly")
			}
		})
	}
}

func TestGetHashesFromTorrent(t *testing.T) {
	t.Run("splits into 20-byte hashes", func(t *testing.T) {
		got, err := gethashesFromtorrent([]byte(twentyA + twentyB))
		if err != nil {
			t.Fatalf("gethashesFromtorrent() failed: %v", err)
		}
		want := [][20]byte{[20]byte([]byte(twentyA)), [20]byte([]byte(twentyB))}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("empty input yields no hashes", func(t *testing.T) {
		got, err := gethashesFromtorrent([]byte{})
		if err != nil {
			t.Fatalf("gethashesFromtorrent() failed: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("got %d hashes, want 0", len(got))
		}
	})

	t.Run("non-multiple of 20 errors", func(t *testing.T) {
		if _, err := gethashesFromtorrent([]byte("not twenty bytes long..")); err == nil {
			t.Error("gethashesFromtorrent() succeeded unexpectedly")
		}
	})
}

func TestParseIntFromData(t *testing.T) {
	data := map[string]any{"n": 7, "s": "str"}

	got, err := parseIntFromData(data, "n")
	if err != nil {
		t.Fatalf("parseIntFromData() failed: %v", err)
	}
	if got != 7 {
		t.Errorf("got %d, want 7", got)
	}

	if _, err := parseIntFromData(data, "s"); err == nil {
		t.Error("parseIntFromData() on a string succeeded unexpectedly")
	}
	if _, err := parseIntFromData(data, "missing"); err == nil {
		t.Error("parseIntFromData() on a missing key succeeded unexpectedly")
	}
}

func TestParseStringFromData(t *testing.T) {
	data := map[string]any{"s": "hello", "n": 7}

	got, err := parseStringFromData(data, "s")
	if err != nil {
		t.Fatalf("parseStringFromData() failed: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}

	if _, err := parseStringFromData(data, "n"); err == nil {
		t.Error("parseStringFromData() on an int succeeded unexpectedly")
	}
	if _, err := parseStringFromData(data, "missing"); err == nil {
		t.Error("parseStringFromData() on a missing key succeeded unexpectedly")
	}
}
