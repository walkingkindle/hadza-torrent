package bencodeparser

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
)

func Decode(location string) (map[string]any, [20]byte, error) {
	bytes, err := openFile(location)
	if err != nil {
		return nil, [20]byte{}, err
	}

	t := &Torrent{bytes, -1, -1, nil}

	value, _, valueErr := t.Dispatch(0)
	if valueErr != nil {
		return nil, [20]byte{}, valueErr
	}

	dict, ok := value.(map[string]any)

	if !ok {
		return nil, [20]byte{}, errors.New("no info dictionary found")
	}

	if t.InfoStart < 0 {
		return nil, [20]byte{}, errors.New("not a torretn, top-level value is not an infoHash")
	}
	if bytes[0] != 'd' {
		return nil, [20]byte{}, errors.New("Unsupported or malformed torrent file")
	}

	infoHash := sha1.Sum(t.Data[t.InfoStart:t.InfoEnd])

	return dict, infoHash, nil
}

func openFile(location string) ([]byte, error) {
	bytes, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("couldn't open %q: %w", location, err)
	}
	return bytes, nil
}
