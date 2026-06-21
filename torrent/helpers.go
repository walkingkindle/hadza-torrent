package torrentparser

import (
	"errors"

	bencodeparser "torrent-client-go/bencode-decoder"
)

func findInfoSlice(data []byte) (int, int, error) {
	if data[0] != 'd' {
		return 0, 0, errors.New("invalid torrent, does not have dictionary in it")
	}

	position := 1

	for position < len(data) && data[position] != 'e' {
		key, pos, err := bencodeparser.ParseString(position, data)
		if err != nil {
			return 0, 0, err
		}

		position = pos

		valueStart := position

		position, err = skipValue(position, data)
		if err != nil {
			return 0, 0, err
		}

		valueEnd := position

		if key == "info" {
			return valueStart, valueEnd, nil
		}
	}

	return 0, 0, errors.New("No info key found in this torrent dictionary")
}

func skipValue(position int, data []byte) (int, error) {
	_, newPos, err := bencodeparser.Dispatch(position, data)

	return newPos, err
}
