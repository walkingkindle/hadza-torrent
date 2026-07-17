package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strings"

	"torrent-client-go/magnet-parser"
	"torrent-client-go/peer"
	torrentParser "torrent-client-go/torrent"
	"torrent-client-go/tracker"
	"torrent-client-go/types"
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
		handleIsTorrentFile(reader, peer)
	default:
		fmt.Println("Invalid input value,bye")
	}
}

func generatePeerID() ([20]byte, error) {
	var bytes [20]byte

	_, err := rand.Read(bytes[:])
	if err != nil {
		return [20]byte{}, err
	}

	return bytes, nil
}

func handleIsTorrentFile(reader *bufio.Reader, peerID [20]byte) {
	fmt.Print("Please input the torrent file location. \n")
	torrentFileLocation, err := readInputFromUser(reader)
	printInputReadErrorIfExists(err)

	fmt.Print("Reading torrent file \n")

	bytes, err := getRawBytesFromFile(torrentFileLocation)

	printInputReadErrorIfExists(err)

	torrent, err := torrentParser.ParseTorrentFile(bytes)

	printInputReadErrorIfExists(err)
	peersResponse, err := tracker.GetPeers(torrent, string(peerID[:]))

	printInputReadErrorIfExists(err)

	conn, err := connectToPeer(peersResponse, torrent, peerID)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	fmt.Printf("Connected with a peer \n")
	fmt.Printf(conn.Peer.IP.String())

	printPeersResponse(peersResponse)
	printInputReadErrorIfExists(err)
}

func connectToPeer(peersResponse []peer.Peer, torrent types.TorrentFile, peerID [20]byte) (connection peer.PeerConnection, err error) {
	for i := 0; i < len(peersResponse); i++ {
		selectedPeer := peersResponse[i]

		conn, err := peer.Connect(selectedPeer, torrent.InfoHash, peerID)
		if err != nil {
			continue
		}

		return conn, nil
	}
	return peer.PeerConnection{}, errors.New("could not establish connection with any of the peers.")
}

func printPeersResponse(peersResponse []peer.Peer) {
	fmt.Printf("%v \n ", len(peersResponse))
	for i := range peersResponse {
		fmt.Printf("Peer response IP,  %s \n", peersResponse[i].IP)
		fmt.Printf("Port, %v \n", peersResponse[i].Port)
	}
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
