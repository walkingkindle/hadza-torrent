// Package types holds various types for torrents;
package types

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	CreatedBy   string
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}
