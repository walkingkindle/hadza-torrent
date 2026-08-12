package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"torrent-client-go/download"
	"torrent-client-go/magnet-parser"
	"torrent-client-go/peer"
	torrentParser "torrent-client-go/torrent"
	"torrent-client-go/tracker"
	"torrent-client-go/types"
)

func main() {
	setupLogging()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Magnet or torrent file. Send 1 or 2. \n")
	fmt.Print("1. Magnet \n")
	fmt.Print("2. Torrent File Location \n")

	sentence, err := readInputFromUser(reader)
	printInputReadErrorIfExists(err)

	peer, err := generatePeerID()
	if err != nil {
		fmt.Print("invalid value for peer id")
		return
	}

	switch sentence {
	case "1":
		handleIsMagnetLink(reader)
	case "2":
		handleIsTorrentFile(reader, peer)
	default:
		fmt.Println("Invalid input value,bye")
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

func generatePeerID() ([20]byte, error) {
	var bytes [20]byte

	_, err := rand.Read(bytes[:])
	if err != nil {
		return [20]byte{}, err
	}

	return bytes, nil
}

func handleIsTorrentFile(reader *bufio.Reader, peerID [20]byte) error {
	fmt.Print("Please input the torrent file location. \n")
	torrentFileLocation, err := readInputFromUser(reader)
	printInputReadErrorIfExists(err)

	fmt.Print("Reading torrent file \n")

	bytes, err := getRawBytesFromFile(torrentFileLocation)

	printInputReadErrorIfExists(err)

	torrent, err := torrentParser.ParseTorrentFile(bytes)

	file, err := createFile(torrent)
	if err != nil {
		return err
	}
	defer file.Close()
	done := make([]bool, len(torrent.PieceHashes))
	printInputReadErrorIfExists(err)

	peers, err := tracker.GetPeers(torrent, string(peerID[:]))
	printInputReadErrorIfExists(err)

	for _, peer := range peers {
		conn, err := connectToPeer(peer, torrent, peerID)
		if err != nil {
			fmt.Printf("could not connect to this peer, %s", err)
			continue
		}
		err = download.Download(conn, torrent, file, done)
		if err == nil {
			break
		}
		fmt.Printf("Error while downloading, %s, retrying", err.Error())
	}

	return nil
}

func createFile(torrent types.TorrentFile) (*os.File, error) {
	file, err := os.Create(torrent.Name)
	if err != nil {
		return nil, err
	}

	if err := file.Truncate(int64(torrent.Length)); err != nil {
		return nil, err
	}
	return file, nil
}

func connectToPeer(selectedPeer peer.Peer, torrent types.TorrentFile, peerID [20]byte) (connection *peer.PeerConnection, err error) {
	conn, err := peer.Connect(selectedPeer, torrent.InfoHash, peerID)
	if err != nil {
		return nil, err
	}

	return &conn, nil
}

func handleIsMagnetLink(reader *bufio.Reader) {
	fmt.Printf("Please gimme magnet link \n")

	sentence, err := readInputFromUser(reader)
	printInputReadErrorIfExists(err)
	result, err := parser.ParseMagnet(sentence)
	printInputReadErrorIfExists(err)
	fmt.Printf("%#v\n", result)
}

func printInputReadErrorIfExists(err error) {
	if err != nil {
		fmt.Print(err.Error())
	}
}

func readInputFromUser(reader *bufio.Reader) (string, error) {
	sentence, err := reader.ReadString('\n')
	if err != nil {
		err = errors.New("error when reading out the sentence")
		return "", err
	}
	cleanedInput := strings.TrimSpace(sentence)
	return cleanedInput, nil
}

func getRawBytesFromFile(location string) ([]byte, error) {
	bytes, err := openFile(location)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func openFile(location string) ([]byte, error) {
	bytes, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("couldn't open %q: %w", location, err)
	}
	return bytes, nil
}
