package download

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"

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
		conn.HandleMessage(msg)
	}

	for i := 0; i < len(torrent.PieceHashes); i++ {
		piece, err := downloadPiece(i, torrent, conn)
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadPiece(i int, torrent types.TorrentFile, conn *peer.PeerConnection) ([]byte, error) {
	buf := make([]byte, torrent.PieceLength)

	filled := 0

	for filled < torrent.PieceLength {
		sendRequest(i, filled, torrent.PieceLength, conn) // handle this error

		for {

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

func sendRequest(i int, filled int, pieceLength int, conn *peer.PeerConnection) {
	msg := peer.Message{ID: 6, Payload: buildPayload(i, filled)}

	serialze := msg.Serialize()

	conn.Conn.Write(serialze)
}

func buildPayload(i int, filled int) []byte {
	pload := make([]byte, 12)

	binary.BigEndian.PutUint32(pload[0:4], uint32(i))

	binary.BigEndian.PutUint32(pload[4:8], uint32(filled))

	binary.BigEndian.PutUint32(pload[8:12], uint32(stdPieceSize))

	return pload
}
