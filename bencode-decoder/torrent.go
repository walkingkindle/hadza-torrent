package bencodeparser

type Torrent struct {
	Data      []byte
	InfoStart int
	InfoEnd   int
	Error     error
}
