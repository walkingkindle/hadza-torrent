package tracker

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"torrent-client-go/types"
)

// bencodePeers wraps raw 6-byte peer entries into a compact tracker response.
func bencodePeers(peerBytes []byte) []byte {
	return []byte("d5:peers" + strconv.Itoa(len(peerBytes)) + ":" + string(peerBytes) + "e")
}

func TestMapPeerBytesToPeer(t *testing.T) {
	t.Run("two peers", func(t *testing.T) {
		// 127.0.0.1:6881 and 10.0.0.5:6882
		peerBytes := []byte{127, 0, 0, 1, 0x1a, 0xe1, 10, 0, 0, 5, 0x1a, 0xe2}
		peers, err := mapPeerBytesToPeer(peerBytes)
		if err != nil {
			t.Fatalf("mapPeerBytesToPeer() failed: %v", err)
		}
		if len(peers) != 2 {
			t.Fatalf("got %d peers, want 2", len(peers))
		}
		if got := peers[0].IP.String(); got != "127.0.0.1" {
			t.Errorf("peers[0].IP = %s, want 127.0.0.1", got)
		}
		if peers[0].Port != 6881 {
			t.Errorf("peers[0].Port = %d, want 6881", peers[0].Port)
		}
		if got := peers[1].IP.String(); got != "10.0.0.5" {
			t.Errorf("peers[1].IP = %s, want 10.0.0.5", got)
		}
		if peers[1].Port != 6882 {
			t.Errorf("peers[1].Port = %d, want 6882", peers[1].Port)
		}
	})

	t.Run("empty yields no peers", func(t *testing.T) {
		peers, err := mapPeerBytesToPeer([]byte{})
		if err != nil {
			t.Fatalf("mapPeerBytesToPeer() failed: %v", err)
		}
		if len(peers) != 0 {
			t.Errorf("got %d peers, want 0", len(peers))
		}
	})

	t.Run("not a multiple of 6 errors", func(t *testing.T) {
		if _, err := mapPeerBytesToPeer([]byte{1, 2, 3, 4, 5}); err == nil {
			t.Error("mapPeerBytesToPeer() succeeded unexpectedly")
		}
	})
}

func TestParseBodyIntoPeerStruct(t *testing.T) {
	t.Run("valid peer response", func(t *testing.T) {
		result := map[string]any{"peers": string([]byte{127, 0, 0, 1, 0x1a, 0xe1})}
		peers, err := parseBodyIntoPeerStruct(result)
		if err != nil {
			t.Fatalf("parseBodyIntoPeerStruct() failed: %v", err)
		}
		if len(peers) != 1 || peers[0].Port != 6881 {
			t.Errorf("unexpected peers: %+v", peers)
		}
	})

	tests := []struct {
		name   string
		result any
	}{
		{name: "not a map", result: "nope"},
		{name: "missing peers key", result: map[string]any{"interval": 900}},
		{name: "peers not a string", result: map[string]any{"peers": 42}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseBodyIntoPeerStruct(tt.result); err == nil {
				t.Error("parseBodyIntoPeerStruct() succeeded unexpectedly")
			}
		})
	}
}

func TestBuildTrackerURL(t *testing.T) {
	infoHash := []byte("12345678901234567890") // 20 bytes
	got, err := buildTrackerURL(infoHash, 1024, "http://tracker.example.com/announce", "peer-id-0123456789ab")
	if err != nil {
		t.Fatalf("buildTrackerURL() failed: %v", err)
	}

	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("returned URL did not parse: %v", err)
	}
	if u.Host != "tracker.example.com" || u.Path != "/announce" {
		t.Errorf("unexpected base URL: %s", got)
	}

	q := u.Query()
	checks := map[string]string{
		"info_hash":  string(infoHash),
		"peer_id":    "peer-id-0123456789ab",
		"port":       "6881",
		"uploaded":   "0",
		"downloaded": "0",
		"compact":    "1",
		"left":       "1024",
	}
	for key, want := range checks {
		if got := q.Get(key); got != want {
			t.Errorf("query %q = %q, want %q", key, got, want)
		}
	}
}

func TestBuildTrackerURLInvalidAnnounce(t *testing.T) {
	if _, err := buildTrackerURL([]byte("hash"), 1, "http://host:notaport", "peer"); err == nil {
		t.Error("buildTrackerURL() succeeded unexpectedly on invalid announce")
	}
}

func TestGetResponse(t *testing.T) {
	t.Run("decodes peers from a 200 response", func(t *testing.T) {
		peerBytes := []byte{127, 0, 0, 1, 0x1a, 0xe1}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(bencodePeers(peerBytes))
		}))
		defer srv.Close()

		peers, err := getResponse(srv.URL)
		if err != nil {
			t.Fatalf("getResponse() failed: %v", err)
		}
		if len(peers) != 1 || peers[0].Port != 6881 {
			t.Errorf("unexpected peers: %+v", peers)
		}
	})

	t.Run("non-200 status errors", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "nope", http.StatusInternalServerError)
		}))
		defer srv.Close()

		if _, err := getResponse(srv.URL); err == nil {
			t.Error("getResponse() succeeded unexpectedly on a 500")
		}
	})

	t.Run("non-bencode body errors", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not bencode"))
		}))
		defer srv.Close()

		if _, err := getResponse(srv.URL); err == nil {
			t.Error("getResponse() succeeded unexpectedly on a bad body")
		}
	})
}

func TestGetPeers(t *testing.T) {
	peerBytes := []byte{127, 0, 0, 1, 0x1a, 0xe1}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The tracker request must carry the query buildTrackerURL produced.
		if r.URL.Query().Get("info_hash") == "" {
			t.Error("tracker request missing info_hash query param")
		}
		w.Write(bencodePeers(peerBytes))
	}))
	defer srv.Close()

	torrent := types.TorrentFile{
		Announce: srv.URL,
		Length:   1024,
	}
	peers, err := GetPeers(torrent, "peer-id-0123456789ab")
	if err != nil {
		t.Fatalf("GetPeers() failed: %v", err)
	}
	if len(peers) != 1 || peers[0].Port != 6881 {
		t.Errorf("unexpected peers: %+v", peers)
	}
}
