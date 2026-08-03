package peer

import (
	"encoding/binary"
	"errors"
)

type PieceBlock struct {
	Index uint32
	Begin uint32
	Data  []byte
}

func ParsePiece(payload []byte) (PieceBlock, error) {
	if len(payload) < 8 {
		return PieceBlock{}, errors.New("malformed payload,returning")
	}

	index := binary.BigEndian.Uint32(payload[0:4])

	begin := binary.BigEndian.Uint32(payload[4:8])

	data := payload[8:]

	return PieceBlock{Index: index, Begin: begin, Data: data}, nil
}
