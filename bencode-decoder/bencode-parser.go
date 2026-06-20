// Package bencodeparser  parses raw bytes into a bencoded format (human-readable.)
package bencodeparser

import (
	"crypto/sha1"
	"errors"
)

func Decode(bytes []byte) (map[string]any, [20]byte, error) {
	t := &Torrent{bytes, -1, -1, nil}
	// TODO: Move this file into a separate implementation. This is a bencode torrent parser, not a generic bencode pareser. make this more SOLID.
	value, _, valueErr := t.Dispatch(0)
	if valueErr != nil {
		return nil, [20]byte{}, valueErr
	}

	dict, ok := value.(map[string]any)

	if !ok {
		return nil, [20]byte{}, errors.New("no info dictionary found")
	}

	if t.InfoStart < 0 {
		return nil, [20]byte{}, errors.New("not a torrent, top-level value is not an infoHash")
	}
	if bytes[0] != 'd' {
		return nil, [20]byte{}, errors.New("unsupported or malformed torrent file")
	}

	infoHash := sha1.Sum(t.Data[t.InfoStart:t.InfoEnd])

	return dict, infoHash, nil
}
