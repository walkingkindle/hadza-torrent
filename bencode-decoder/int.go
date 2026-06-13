package bencodeparser

import (
	"errors"
	"strconv"
)

func (t *Torrent) ParseInt(startPosition int) (int, int, error) {
	if startPosition < 0 || startPosition >= len(t.Data) {
		return 0, 0, errors.New("start position out of bounds")
	}
	if t.Data[startPosition] != 'i' {
		return 0, 0, errors.New("invalid type, does not start with correct letter")
	}
	rel := indexOf(t.Data[startPosition:], 'e')

	if rel == -1 {
		return 0, 0, errors.New("invalid index of end, malformed error")
	}

	absolutePositionEnd := startPosition + rel
	innerBytes := t.Data[startPosition+1 : absolutePositionEnd]

	err := validateInnerBytes(innerBytes)
	if err != nil {
		return 0, 0, err
	}
	stringVal, err := strconv.Atoi(string(innerBytes))
	if err != nil {
		return 0, 0, err
	}

	return int(stringVal), absolutePositionEnd + 1, nil
}

func validateInnerBytes(innerBytes []byte) error {
	if len(innerBytes) == 0 {
		return errors.New("empty integer body")
	}
	if len(innerBytes) > 1 && innerBytes[0] == '0' {
		return errors.New("leading zeros are forbidden")
	}
	if len(innerBytes) >= 2 && innerBytes[0] == '-' && innerBytes[1] == '0' {
		return errors.New("negative zero is forbidden")
	}

	return nil
}
