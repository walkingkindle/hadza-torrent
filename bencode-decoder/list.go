package bencodeparser

import (
	"errors"
)

func ParseList(startPosition int, data []byte) ([]any, int, error) {
	if startPosition < 0 || startPosition >= len(data) {
		return []any{}, 0, errors.New("start position out of bounds")
	}
	if data[startPosition] != 'l' {
		return []any{}, 0, errors.New("invalid type, does not start with correct letter")
	}
	startPosition++

	result := []any{}
	for startPosition < len(data) && data[startPosition] != 'e' {
		value, next, err := Dispatch(startPosition, data)
		if err != nil {
			return nil, 0, err
		}

		result = append(result, value)

		startPosition = next
	}

	if startPosition >= len(data) {
		return nil, 0, errors.New("malformed list: ran out of bytes before finding 'e'")
	}

	return result, startPosition + 1, nil
}
