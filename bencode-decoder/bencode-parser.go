// Package bencodeparser  parses raw bytes into a bencoded format (human-readable.)
package bencodeparser

func Decode(data []byte) (any, error) {
	val, _, err := Dispatch(0, data)

	return val, err
}
