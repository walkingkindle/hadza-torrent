package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"torrent-client-go/session"
)

func main() {
	// TODO: move this stupid logger out of main
	setupLogging()

	location, err := ParseLocationFromArgs()
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
	err = session.DownloadFileFromTorrent(location)
	if err != nil {
		fmt.Printf("error downloading file, %s, /n", err.Error())
		os.Exit(1)
	}
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
