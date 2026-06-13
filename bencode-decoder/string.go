package bencodeparser

import (
	"errors"
	"strconv"
)

func (t *Torrent) ParseString(position int) (string, int, error) {
	if position < 0 || position >= len(t.Data) {
		return "", 0, errors.New("start position out of bounds")
	}
	b := t.Data[position]
	if !isDigit(b) {
		return "", 0, errors.New("string must start with a digit")
	}
	length, newPosition, err := getLength(position, t.Data)
	if err != nil {
		return "", -1, err
	}
	position = newPosition
	value, newPosition, err := getStringValue(position, length, t.Data)
	if err != nil {
		return "", -1, err
	}

	return value, newPosition, nil
}

func getStringValue(position int, length int, byteArr []byte) (string, int, error) {
	if (position + length) <= len(byteArr) {
		value := byteArr[position:(position + length)]
		return string(value), (position + length), nil
	}
	return "", 0, errors.New("malformed string value")
}

func getLength(lengthStartPosition int, bytes []byte) (int, int, error) {
	rel := indexOf(bytes[lengthStartPosition:], ':')

	if rel == -1 {
		return -1, -1, errors.New("no colon found in string length")
	}
	absoluteColonIndx := lengthStartPosition + rel
	keyLength, err := strconv.Atoi(string(bytes[lengthStartPosition:absoluteColonIndx]))
	if err != nil {
		return -1, -1, errors.New("conversion failed")
	}

	return keyLength, absoluteColonIndx + 1, nil
}
