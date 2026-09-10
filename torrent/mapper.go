package torrentparser

import (
	"errors"
	"fmt"

	"torrent-client-go/types"
)

func mapInfoToTorrentFile(
	info map[string]any,
	infoHash [20]byte,
) (types.TorrentFile, error) {
	var torrent types.TorrentFile
	torrent.InfoHash = infoHash

	name, err := parseStringFromData(info, "name")
	if err != nil {
		return types.TorrentFile{}, err
	}
	torrent.Name = name

	pieceLength, err := parseIntFromData(info, "piece length")
	if err != nil {
		return types.TorrentFile{}, err
	}
	torrent.PieceLength = pieceLength

	if filesValue, exists := info["files"]; exists {
		files, err := parseTorrentFiles(filesValue)
		if err != nil {
			return types.TorrentFile{}, err
		}

		torrent.Files = files

		for _, file := range files {
			torrent.Length += int(file.Length)
		}
	} else {
		length, err := parseIntFromData(info, "length")
		if err != nil {
			return types.TorrentFile{}, err
		}
		torrent.Length = length

		torrent.Files = []types.TorrentFileEntry{
			{
				Path:   []string{torrent.Name},
				Length: int64(length),
			},
		}
	}

	piecesStr, ok := info["pieces"].(string)
	if !ok {
		return types.TorrentFile{}, errors.New("hashes collection invalid")
	}

	hashesCollection, err := gethashesFromtorrent([]byte(piecesStr))
	if err != nil {
		return types.TorrentFile{}, err
	}

	torrent.PieceHashes = hashesCollection

	return torrent, nil
}

func parseTorrentFiles(value any) ([]types.TorrentFileEntry, error) {
	files, ok := value.([]any)
	if !ok {
		return nil, errors.New("invalid files collection")
	}

	result := make([]types.TorrentFileEntry, 0, len(files))

	for i, value := range files {
		fileData, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid file entry at index %d", i)
		}

		length, err := parseIntFromData(fileData, "length")
		if err != nil {
			return nil, fmt.Errorf(
				"invalid file %d length: %w",
				i,
				err,
			)
		}

		pathValue, ok := fileData["path"]
		if !ok {
			return nil, fmt.Errorf("file %d missing path", i)
		}

		pathParts, err := parsePath(pathValue)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid path for file %d: %w",
				i,
				err,
			)
		}

		result = append(result, types.TorrentFileEntry{
			Path:   pathParts,
			Length: int64(length),
		})
	}

	return result, nil
}

func parsePath(value any) ([]string, error) {
	path, ok := value.([]any)
	if !ok {
		return nil, errors.New("path is not a list")
	}

	result := make([]string, 0, len(path))

	for i, part := range path {
		str, ok := part.(string)
		if !ok {
			return nil, fmt.Errorf(
				"path component %d is not a string",
				i,
			)
		}

		result = append(result, str)
	}

	return result, nil
}

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
	if filesValue, exists := info["files"]; exists {
		files, err := parseTorrentFiles(filesValue)
		if err != nil {
			return types.TorrentFile{}, err
		}

		torrent.Files = files
	} else {
		length, err := parseIntFromData(info, "length")
		if err != nil {
			return types.TorrentFile{}, err
		}
		torrent.Length = length

		torrent.Files = []types.TorrentFileEntry{
			{
				Path:   []string{torrent.Name},
				Length: int64(length),
			},
		}
	}

	// TODO: Handle case when you get an announce-list here instead of announce and we already have the concurrecnyl to fetch all the peers at the same time lol
	announce, err := parseStringFromData(data, "announce")
	if err == nil {
		torrent.Announce = announce
	}
	pieceLength, err := parseIntFromData(info, "piece length")
	if err != nil {
		return types.TorrentFile{}, err
	}

	torrent.PieceLength = pieceLength

	createdBy, err := parseStringFromData(data, "created by")
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
	value, ok := data[key]
	if !ok {
		return 0, fmt.Errorf("missing %s property", key)
	}

	switch number := value.(type) {
	case int:
		return number, nil
	case int64:
		return int(number), nil
	default:
		return 0, fmt.Errorf(
			"cannot convert %s property to int, got %T",
			key,
			value,
		)
	}
}

func parseStringFromData(data map[string]any, key string) (string, error) {
	str, ok := data[key].(string)

	if !ok {
		return "", fmt.Errorf("cannot convert the property to string, %s, failed", key)
	}

	return str, nil
}
