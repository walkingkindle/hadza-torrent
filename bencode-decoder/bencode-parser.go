package bencodeparser

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
)

var torrentBytes = []byte{}

func Decode(location string) ([]byte, error) {
	bytes, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("could not open the file %q: %w", location, err)
	}

	torrentBytes = bytes

	position := 0
	for i := 0; i < len(torrentBytes); {
		switch bytes[position] {
		case 'd':
			return parsedict(position)
		}
	}

	return bytes, nil
}

func parsedict(position int) ([]byte, error) {
	if torrentBytes[position] != 'd' {
		return nil, fmt.Errorf("not a dictionary at position %d", position)
	}

	lengthStartPosition := position + 1

	length, absoluteColonIdx, err := getLength(lengthStartPosition)
	if err != nil {
		fmt.Printf("formatting error, %s", err)
	}

	key, keyEndPosition := extractKey(absoluteColonIdx, length)

	fmt.Printf("key and position are %s, %d", key, torrentBytes[keyEndPosition])

	newLength, _, _ := getLength(keyEndPosition)

	fmt.Printf("%d \n", newLength)

	key2, _ := extractKey(newLength, newLength)

	fmt.Printf("%s \n", key2)

	return nil, nil
}

func extractKey(absoluteColonIdx int, length int) (string, int) {
	stringDataStart := absoluteColonIdx + 1
	stringDataEnd := stringDataStart + length
	stringValue := torrentBytes[stringDataStart:stringDataEnd]
	return string(stringValue), stringDataEnd
}

func getLength(lengthStartPosition int) (int, int, error) {
	relativevColonIndx := bytes.IndexByte(torrentBytes[lengthStartPosition:], ':')
	if relativevColonIndx == -1 {
		return 0, 0, fmt.Errorf("colon not found")
	}
	absoluteColonIdx := lengthStartPosition + relativevColonIndx

	stringLengthBytes := torrentBytes[lengthStartPosition:absoluteColonIdx]

	length, err := strconv.Atoi(string(stringLengthBytes))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid string length, %s", err)
	}
	return length, absoluteColonIdx, nil
}
