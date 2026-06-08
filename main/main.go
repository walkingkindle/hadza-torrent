package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	bencodeparser "torrent-client-go/bencode-decoder"
	"torrent-client-go/magnet-parser"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Magnet or torrent file. Send 1 or 2. \n")
	fmt.Print("1. Magnet \n")
	fmt.Print("2. Torrent File Location \n")

	sentence, err := readInputFromUser(reader)
	printInputReadErrorIfExists(err)

	switch sentence {
	case "1":
		handleIsMagnetLink(reader)
	case "2":
		handleIsTorrentFile(reader)
	default:
		fmt.Println("Invalid input value,bye")
	}
}

func handleIsTorrentFile(reader *bufio.Reader) {
	fmt.Print("Please input the torrent file location. \n")
	torrentFileLocation, err := readInputFromUser(reader)
	printInputReadErrorIfExists(err)

	result, err := bencodeparser.Decode(torrentFileLocation)

	printInputReadErrorIfExists(err)

	fmt.Printf("raw bytes: %v\n", result)
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
