// Package bencodeparser  parses raw bytes into a bencoded format (human-readable.)
package bencodeparser

import (
	"bytes"
	"errors"
	"sort"
	"strconv"
)

func Decode(data []byte) (any, error) {
	val, _, err := Dispatch(0, data)

	return val, err
}

func Encode(data any) ([]byte, error) {
	if data == nil {
		return nil, errors.New("data cannot be blank")
	}
	switch data := data.(type) {
	case int:
		return []byte("i" + strconv.Itoa(data) + "e"), nil
	case string:
		return []byte(strconv.Itoa(len(data)) + ":" + data), nil
	case []byte:
		return []byte(strconv.Itoa(len(data)) + ":" + string(data)), nil
	case []any:
		var buf bytes.Buffer
		buf.WriteByte('l')
		for _, item := range data {
			itemBytes, err := Encode(item)
			if err != nil {
				return nil, err
			}
			buf.Write(itemBytes)
		}
		buf.WriteByte('e')
		return buf.Bytes(), nil
	case map[string]any:
		keys := make([]string, 0, len(data))
		for k := range data {
			keys = append(keys, k)
		}
		sort.Strings(keys) // bencode requires sorted keys

		var buf bytes.Buffer
		buf.WriteByte('d')
		for _, k := range keys {
			keyBytes, err := Encode(k)
			if err != nil {
				return nil, err
			}
			valBytes, err := Encode(data[k])
			if err != nil {
				return nil, err
			}
			buf.Write(keyBytes)
			buf.Write(valBytes)
		}
		buf.WriteByte('e')
		return buf.Bytes(), nil
	}

	return nil, errors.New("unsupported bencode type")
}
