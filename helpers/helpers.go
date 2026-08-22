// Package helpers are just some helper functions
package helpers

import "crypto/rand"

func CountTrue(bools []bool) int {
	count := 0

	for _, b := range bools {
		if b {
			count++
		}
	}

	return count
}

func GeneratePeerID() ([20]byte, error) {
	var bytes [20]byte
	_, err := rand.Read(bytes[:])
	if err != nil {
		return [20]byte{}, err
	}

	return bytes, nil
}
