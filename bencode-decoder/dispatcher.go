package bencodeparser

import "errors"

func Dispatch(position int, data []byte) (any, int, error) {
	if position < 0 || position >= len(data) {
		return nil, 0, errors.New("start position out of bounds")
	}
	b := data[position]
	switch {
	case b == 'i':
		return ParseInt(position, data)
	case b == 'd':
		return ParseDict(position, data)
	case isDigit(b):
		return ParseString(position, data)
	case b == 'l':
		return ParseList(position, data)
	default:
		return nil, 0, errors.New("torrent malformed")
	}
}
