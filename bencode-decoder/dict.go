package bencodeparser

import "errors"

func (t *Torrent) ParseDict(position int) (map[string]any, int, error) {
	if position < 0 || position >= len(t.Data) {
		return nil, 0, errors.New("start position out of bounds")
	}
	if t.Data[position] != 'd' {
		return nil, 0, errors.New("value is not a dict, returning")
	}
	position++
	m := make(map[string]any)
	for position < len(t.Data) && t.Data[position] != 'e' {
		key, pos, keyErr := t.ParseString(position)
		if keyErr != nil {
			return nil, 0, keyErr
		}
		position = pos
		valueStart := position
		value, pos, valErr := t.Dispatch(position)
		if valErr != nil {
			return nil, 0, valErr
		}

		if key == "info" {
			t.InfoStart = valueStart
			t.InfoEnd = pos
		}

		position = pos

		m[key] = value

	}

	if position >= len(t.Data) {
		return nil, 0, errors.New("undeterminated dict: ran out of bytes before 'e'")
	}

	return m, position + 1, nil
}
