package download

import (
	"crypto/sha1"
	"fmt"

	"torrent-client-go/types"
)

func GetPieceLength(i int, torrent types.TorrentFile) int {
	pieceStart := i * torrent.PieceLength

	pieceEnd := min(pieceStart+torrent.PieceLength, torrent.Length)
	thisPieceLength := pieceEnd - pieceStart

	return thisPieceLength
}

func VerifyPiece(piece []byte, expected [20]byte) error {
	actual := sha1.Sum(piece)

	if actual != expected {
		return fmt.Errorf("piece hash mismatch: got %x, want %x", actual[:8], expected[:8])
	}

	return nil
}
