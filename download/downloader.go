// download package handles the download feed and writes to file
package download

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"os"
	"time"

	"torrent-client-go/peer"
	"torrent-client-go/types"
)

const stdPieceSize = 16384 // 16kib

func Download(conn *peer.PeerConnection, torrent types.TorrentFile) error {
	for conn.Choked {
		msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		err = conn.HandleMessage(msg)
	}

	file, err := os.Create(torrent.Name)
	if err != nil {
		return errors.New("system error could not create the file")
	}

	defer file.Close()

	for i := 0; i < len(torrent.PieceHashes); i++ {
		piece, err := downloadPiece(i, torrent, conn)
		if err != nil {
			return err
		}
		_, err = file.WriteAt(piece, int64(i*torrent.PieceLength))
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadPiece(i int, torrent types.TorrentFile, conn *peer.PeerConnection) ([]byte, error) {
	pieceLength := getPieceLength(i, torrent)
	buf := make([]byte, pieceLength)

	filled := 0

	for filled < pieceLength {
		err := sendRequest(i, filled, pieceLength, conn) // handle this error
		if err != nil {
			return nil, err
		}
		for {
			conn.Conn.SetDeadline(time.Now().Add(6))
			mes, err := conn.ReadMessage()
			if err != nil {
				return nil, err
			}
			block, err := conn.HandleMessage(mes)
			if err != nil {
				return nil, err
			}

			if block != nil {
				copy(buf[block.Begin:], block.Data)
				filled += len(block.Data)
				break
			}
		}
	}

	hash := sha1.Sum(buf)

	if hash != torrent.PieceHashes[i] {
		return nil, errors.New("piece hash mismatch")
	}

	return buf, nil
}

func getPieceLength(i int, torrent types.TorrentFile) int {
	pieceStart := i * torrent.PieceLength

	pieceEnd := pieceStart + torrent.PieceLength

	if pieceEnd > torrent.Length {
		pieceEnd = torrent.Length
	}
	thisPieceLength := pieceEnd - pieceStart

	return thisPieceLength
}

func sendRequest(i int, filled int, pieceLength int, conn *peer.PeerConnection) error {
	msg := peer.Message{ID: 6, Payload: buildPayload(i, filled, pieceLength)}

	serialze := msg.Serialize()

	_, err := conn.Conn.Write(serialze)
	if err != nil {
		return err
	}

	return nil
}

func buildPayload(i int, filled int, pieceLength int) []byte {
	pload := make([]byte, 12)

	binary.BigEndian.PutUint32(pload[0:4], uint32(i))

	binary.BigEndian.PutUint32(pload[4:8], uint32(filled))

	binary.BigEndian.PutUint32(pload[8:12], uint32(min(stdPieceSize, pieceLength-filled)))

	return pload
}
