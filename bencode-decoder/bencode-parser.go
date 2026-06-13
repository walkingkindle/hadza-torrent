package bencodeparser

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

func Decode(location string) ([]byte, error) {
	bytes, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("could nt open the file %q: %w", location, err)
	}

	position := 0
	for i := 0; i < len(bytes); {
		switch bytes[position] {
		case 'd':
			return parsedict(position, bytes)
		}
	}

	return bytes, nil
}

func parsedict(position int, bytesArr []byte) ([]byte, error) {
	if bytesArr[position] != 'd' {
		return nil, fmt.Errorf("not a dctionary at position %d", position)
	}
	position++

	m := make(map[string]string)
	for bytesArr[position] != 'e' {
		key, pos, keyErr := parseString(position, bytesArr)
		fmt.Printf("\n key is %s \n", key)
		fmt.Printf("pos is %d", pos)
		position = pos

		if keyErr != nil {
			return bytesArr, keyErr
		}

		value, pos, err := parseString(position, bytesArr)

		if err != nil {
			return bytesArr, err
		}

		fmt.Printf("\n value is %s \n", value)
		fmt.Printf("pos is %d", pos)
		position = pos

		m[key] = value

	}

	return nil, nil
}

func parseString(position int, bytes []byte) (string, int, error) {
	length, newPosition, err := getLength(position, bytes)
	if err != nil {
		return "", -1, err
	}

	position = newPosition

	value, newPosition := getStringValue(position, length, bytes)

	return value, newPosition, nil
}

func getStringValue(position int, length int, byteArr []byte) (string, int) {
	value := byteArr[position:(position + length)]

	return string(value), (position + length)
}

func getLength(lengthStartPosition int, bytesArr []byte) (int, int, error) {
	lengthEndPosition := indexOf(bytesArr[lengthStartPosition:], ':')
	absoluteColonIndx := lengthStartPosition + lengthEndPosition
	keyLength, err := strconv.Atoi(string(bytesArr[lengthStartPosition:absoluteColonIndx]))
	if err != nil {
		fmt.Print("Error with conversion to int")

		return -1, -1, errors.New("Conversion failed")
	}

	return keyLength, absoluteColonIndx + 1, nil
}

func indexOf(haystack []byte, needle byte) int {
	for i, v := range haystack {
		if v == needle {
			return i
		}
	}
	return -1
}
