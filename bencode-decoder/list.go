package bencodeparser

import (
	"errors"
)

func (t *Torrent) ParseList(startPosition int) ([]any, int, error) {
	if startPosition < 0 || startPosition >= len(t.Data) {
		return []any{}, 0, errors.New("start position out of bounds")
	}
	if t.Data[startPosition] != 'l' {
		return []any{}, 0, errors.New("invalid type, does not start with correct letter")
	}
	startPosition++

	result := []any{}
	for startPosition < len(t.Data) && t.Data[startPosition] != 'e' {
		value, next, err := t.Dispatch(startPosition)
		if err != nil {
			return nil, 0, err
		}

		result = append(result, value)

		startPosition = next
	}

	if startPosition >= len(t.Data) {
		return nil, 0, errors.New("malformed list: ran out of bytes before finding 'e'")
	}

	return result, startPosition + 1, nil
}
