package bencodeparser

import "errors"

func (t *Torrent) Dispatch(position int) (any, int, error) {
	if position < 0 || position >= len(t.Data) {
		return nil, 0, errors.New("start position out of bounds")
	}
	b := t.Data[position]
	switch {
	case b == 'i':
		return t.ParseInt(position)
	case b == 'd':
		return t.ParseDict(position)
	case isDigit(b):
		return t.ParseString(position)
	case b == 'l':
		return t.ParseList(position)
	default:
		return nil, 0, errors.New("torrent malformed")
	}
}
