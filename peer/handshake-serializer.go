package peer

import (
	"bytes"
	"fmt"
)

const protocol = "BitTorrent protocol"

func (h Handshake) Serialize() ([68]byte, error) {
	var arr [68]byte

	// The buffer writes into arr's backing storage. This is safe only because
	// the writes below total exactly len(arr) bytes and never force a realloc.
	buf := bytes.NewBuffer(arr[:0])

	buf.WriteByte(byte(len(protocol))) // pstrlen; bytes.Buffer writes never error
	buf.WriteString(protocol)          // pstr
	buf.Write(make([]byte, 8))         // reserved
	buf.Write(h.InfoHash[:])
	buf.Write(h.PeerID[:])

	if buf.Len() != len(arr) {
		return [68]byte{}, fmt.Errorf("peer: serialized handshake is %d bytes, want %d", buf.Len(), len(arr))
	}

	return arr, nil
}

// func ParseHandshake(data []byte) (h Handshake, err error)
