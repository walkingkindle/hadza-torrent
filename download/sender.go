package download

import (
	"encoding/binary"
	"log/slog"

	"torrent-client-go/peer"
)

func SendInterested(i int, filled int, pieceLength int, conn *peer.PeerConnection) error {
	return sendRequest(i, filled, pieceLength, conn)
}

func sendRequest(
	pieceIndex int,
	begin int,
	length int,
	conn *peer.PeerConnection,
) error {
	slog.Debug(
		"requesting block",
		"piece", pieceIndex,
		"begin", begin,
		"length", length,
	)

	return conn.Send(peer.Message{
		ID:      peer.MsgRequest,
		Payload: buildPayload(pieceIndex, begin, length),
	})
}

func buildPayload(
	pieceIndex int,
	begin int,
	length int,
) []byte {
	payload := make([]byte, 12)

	binary.BigEndian.PutUint32(
		payload[0:4],
		uint32(pieceIndex),
	)

	binary.BigEndian.PutUint32(
		payload[4:8],
		uint32(begin),
	)

	binary.BigEndian.PutUint32(
		payload[8:12],
		uint32(length),
	)

	return payload
}
