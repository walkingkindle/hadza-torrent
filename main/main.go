package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go/v5"

	"torrent-client-go/download"
	"torrent-client-go/helpers"
	"torrent-client-go/peer"
	"torrent-client-go/session"
	"torrent-client-go/tracker"
	"torrent-client-go/types"
)

func main() {
	// TODO: move this stupid logger out of main
	setupLogging()

	location, err := ParseLocationFromArgs()
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	session.DownloadFileFromTorrent(location)
}

// setupLogging sends logs to stderr so they stay separate from the prompts on
// stdout. TORRENT_LOG=debug adds the per-message wire trace: every message
// sent and received, every block, every request.
func setupLogging() {
	level := slog.LevelInfo

	switch strings.ToLower(os.Getenv("TORRENT_LOG")) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().Format("15:04:05.000"))
			}
			return a
		},
	})

	slog.SetDefault(slog.New(handler))
}

func connectToPeer(selectedPeer peer.Peer, torrent types.TorrentFile, peerID [20]byte) (connection *peer.PeerConnection, err error) {
	conn, err := peer.Connect(selectedPeer, torrent.InfoHash, peerID)
	if err != nil {
		return nil, err
	}

	return &conn, nil
}
