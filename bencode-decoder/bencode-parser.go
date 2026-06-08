package bencodeparser

import (
	"fmt"
	"os"
)

func Decode(location string) ([]byte, error) {
	bytes, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("could not open the file %q: %w", location, err)
	}

	return bytes, nil
}
