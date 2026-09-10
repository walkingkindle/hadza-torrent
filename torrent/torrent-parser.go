// Package torrentparser parses torrent files from a dedicated map[string]any data and its infohash
package torrentparser

import (
	"crypto/sha1"
	"errors"
	"fmt"

	"torrent-client-go/file"
	"torrent-client-go/types"
)

func ParseTorrentFile(bencoded file.BencodedTorrent) (types.TorrentFile, error) {
	infoStart, infoEnd, err := findInfoSlice(bencoded.FileBytes)
	if err != nil {
		return types.TorrentFile{}, err
	}
	if bencoded.FileBytes[0] != 'd' {
		return types.TorrentFile{}, errors.New("unsupported or malformed torrent file")
	}

	infoHash := sha1.Sum(bencoded.FileBytes[infoStart:infoEnd])

	torrentFile, err := mapDataToTorrentFile(bencoded.Dict, infoHash)
	if err != nil {
		return types.TorrentFile{}, err
	}
	fmt.Printf("%+v\n", torrentFile)

	return torrentFile, nil
}

func ParseInfoMetadata(
	infoDict map[string]any,
	infoHash [20]byte,
) (types.TorrentFile, error) {
	return mapInfoToTorrentFile(infoDict, infoHash)
}
