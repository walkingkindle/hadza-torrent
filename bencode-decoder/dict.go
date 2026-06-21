package bencodeparser

import "errors"

func ParseDict(position int, data []byte) (map[string]any, int, error) {
	if position < 0 || position >= len(data) {
		return nil, 0, errors.New("start position out of bounds")
	}
	if data[position] != 'd' {
		return nil, 0, errors.New("value is not a dict, returning")
	}
	position++
	m := make(map[string]any)
	for position < len(data) && data[position] != 'e' {
		key, pos, keyErr := ParseString(position, data)
		if keyErr != nil {
			return nil, 0, keyErr
		}
		position = pos
		value, pos, valErr := Dispatch(position, data)
		if valErr != nil {
			return nil, 0, valErr
		}

		position = pos

		m[key] = value
	}

	if position >= len(data) {
		return nil, 0, errors.New("undeterminated dict: ran out of bytes before 'e'")
	}

	return m, position + 1, nil
}
