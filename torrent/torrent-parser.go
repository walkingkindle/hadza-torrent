// Package torrentparser parses torrent files from a dedicated map[string]any data and its infohash
package torrentparser

import (
	"crypto/sha1"
	"errors"

	bencodeparser "torrent-client-go/bencode-decoder"
	"torrent-client-go/types"
)

func ParseTorrentFile(bytes []byte) (types.TorrentFile, error) {
	value, valueErr := bencodeparser.Decode(bytes)
	if valueErr != nil {
		return types.TorrentFile{}, valueErr
	}

	dict, ok := value.(map[string]any)

	if !ok {
		return types.TorrentFile{}, errors.New("not a valid torrent")
	}

	infoStart, infoEnd, err := findInfoSlice(bytes)
	if err != nil {
		return types.TorrentFile{}, err
	}
	if bytes[0] != 'd' {
		return types.TorrentFile{}, errors.New("unsupported or malformed torrent file")
	}

	infoHash := sha1.Sum(bytes[infoStart:infoEnd])

	torrentFile, err := mapDataToTorrentFile(dict, infoHash)
	if err != nil {
		return types.TorrentFile{}, err
	}

	return torrentFile, nil
}

// func printInfoDict(info map[string]any) {
// 	for key, value := range info {
// 		if key != "pieces" {
// 			fmt.Printf("Key: %s, Value : %v", key, value)
// 		}
// 	}
// }
