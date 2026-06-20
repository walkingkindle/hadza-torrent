package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strings"

	bencodeparser "torrent-client-go/bencode-decoder"
	"torrent-client-go/magnet-parser"
	torrentParser "torrent-client-go/torrent"
	"torrent-client-go/tracker"
)

func main() {
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
		handleIsTorrentFile(reader, string(peer))
	default:
		fmt.Println("Invalid input value,bye")
	}
}

func generatePeerID() ([]byte, error) {
	bytes := make([]byte, 20)

	_, err := rand.Read(bytes)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

func handleIsTorrentFile(reader *bufio.Reader, peerID string) {
	fmt.Print("Please input the torrent file location. \n")
	torrentFileLocation, err := readInputFromUser(reader)
	printInputReadErrorIfExists(err)

	fmt.Print("Reading torrent file \n")

	bytes, err := getRawBytesFromFile(torrentFileLocation)

	printInputReadErrorIfExists(err)

	dict, infoHash, err := bencodeparser.Decode(bytes)

	printInputReadErrorIfExists(err)

	torrentStruct, err := torrentParser.MapDataToTorrentFile(dict, infoHash)

	printInputReadErrorIfExists(err)
	url, err := tracker.GetPeers(torrentStruct, peerID)

	fmt.Print(url)
	printInputReadErrorIfExists(err)
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
