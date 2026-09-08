// Package file is used for parsing the (torrent) file from a specific location
package file

import (
	"errors"
	"fmt"
	"os"

	bencodeparser "torrent-client-go/bencode-decoder"
)

func GetTorrentMap(location string) (BencodedTorrent, error) {
	bytes, err := getRawBytesFromFile(location)
	if err != nil {
		return BencodedTorrent{}, err
	}

	decoded, err := bencodeparser.Decode(bytes)
	if err != nil {
		return BencodedTorrent{}, err
	}

	dict, ok := decoded.(map[string]any)

	if !ok {
		return BencodedTorrent{}, errors.New("not a valid torrent")
	}

	return BencodedTorrent{
		FileBytes: bytes,
		Dict:      dict,
	}, nil
}

func getRawBytesFromFile(location string) ([]byte, error) {
	bytes, err := openFile(location)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

type BencodedTorrent struct {
	FileBytes []byte
	Dict      map[string]any
}

func openFile(location string) ([]byte, error) {
	bytes, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("couldn't open %q: %w", location, err)
	}
	return bytes, nil
}
