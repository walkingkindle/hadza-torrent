package parser

import (
	"testing"
)

func TestStringReceivedIsMagnet(t *testing.T) {
	magnetLink := "magnet:?xt=urn:btih:c12fe1c06bba254a9dc9f519b335aa7c1367a88a&dn=My+File&tr=udp%3A%2F%2Ftracker.example.com%3A6969"

	result := isAMagnet(magnetLink)

	if !result {
		t.Errorf("Magnet expected to be true for the string valid magnet, returned %t", result)
	}
}

func TestStringReceivedIsNotMagnet(t *testing.T) {
	magnetLink := "whatever"

	result := isAMagnet(magnetLink)

	if result {
		t.Errorf("Expected false but returned %t", result)
	}
}
