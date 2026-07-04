package peer

import (
	"bytes"
	"errors"
	"fmt"
)

const protocol = "BitTorrent protocol"

const handshakeLength = 68

func (h Handshake) Serialize() ([handshakeLength]byte, error) {
	var arr [handshakeLength]byte

	buf := bytes.NewBuffer(arr[:0])

	buf.WriteByte(byte(len(protocol))) // pstrlen; bytes.Buffer writes never error
	buf.WriteString(protocol)          // pstr
	buf.Write(make([]byte, 8))         // reserved
	buf.Write(h.InfoHash[:])
	buf.Write(h.PeerID[:])

	if buf.Len() != len(arr) {
		return [handshakeLength]byte{}, fmt.Errorf("peer: serialized handshake is %d bytes, want %d", buf.Len(), len(arr))
	}

	return arr, nil
}

func ParseHandshake(data []byte) (h Handshake, err error) {
	if len(data) < handshakeLength {
		return h, errors.New("malformed data array, data must be 68 bytes")
	}
	if data[0] != byte(len(protocol)) {
		return h, errors.New("not a bittorrent message or wrong protocol")
	}
	if string(data[1:1+len(protocol)]) != protocol {
		return h, errors.New("not a bittorrent message or wrong protocol")
	}

	copy(h.InfoHash[:], data[28:48])

	copy(h.PeerID[:], data[48:68])

	return h, nil
}
