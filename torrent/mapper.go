package torrentparser

import (
	"errors"
	"fmt"

	"torrent-client-go/types"
)

func mapDataToTorrentFile(data map[string]any, infohash [20]byte) (torrent types.TorrentFile, err error) {
	torrent.InfoHash = infohash

	info, ok := data["info"].(map[string]any)

	if !ok {
		return types.TorrentFile{}, errors.New("invalid infohash table")
	}
	name, err := parseStringFromData(info, "name")
	if err != nil {
		return types.TorrentFile{}, err
	}
	torrent.Name = name

	announce, err := parseStringFromData(data, "announce")
	if err != nil {
		return types.TorrentFile{}, err
	}
	torrent.Announce = announce

	pieceLength, err := parseIntFromData(info, "piece length")
	if err != nil {
		return types.TorrentFile{}, err
	}

	torrent.PieceLength = pieceLength

	length, err := parseIntFromData(info, "length")
	if err != nil {
		return types.TorrentFile{}, err
	}

	torrent.Length = length

	createdBy, err := parseStringFromData(info, "created by")
	if err != nil {
		fmt.Print("Torrent has no created by column, skipping \n")
	}

	torrent.CreatedBy = createdBy

	piecesStr, ok := info["pieces"].(string)

	if !ok {
		return types.TorrentFile{}, errors.New("hashes collection invalid")
	}

	allHashes := []byte(piecesStr)

	hashesCollection, err := gethashesFromtorrent(allHashes)
	// TODO: Maybe here probably is a better way to have hashesCollection in byte right away, change the parser appropriately
	if err != nil {
		return types.TorrentFile{}, err
	}

	torrent.PieceHashes = hashesCollection

	return torrent, nil
}

func gethashesFromtorrent(allHashes []byte) (hashesCollection [][20]byte, err error) {
	if len(allHashes)%20 != 0 {
		return [][20]byte{}, errors.New("malformed hash collection")
	}
	result := [][20]byte{}

	for i := 0; i < len(allHashes); i += 20 {
		arr := allHashes[i : i+20]
		result = append(result, [20]byte(arr))
	}

	return result, nil
}

func parseIntFromData(data map[string]any, key string) (int, error) {
	number, ok := data[key].(int)

	if !ok {
		return 0, errors.New("cannot convert the name property to int, failed")
	}

	return number, nil
}

func parseStringFromData(data map[string]any, key string) (string, error) {
	str, ok := data[key].(string)

	if !ok {
		return "", fmt.Errorf("cannot convert the property to string, %s, failed", key)
	}

	return str, nil
}
