// Package peer tries to handshake with the return peers and the connection struct
package peer

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
	"time"
)

func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}

	length := uint32(len(m.Payload) + 1)

	buf := make([]byte, 4+length)

	binary.BigEndian.PutUint32(buf[0:4], length)

	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)

	return buf
}

// HandleMessage applies msg to the connection state. Only a piece message
// (ID 7) yields a block; every other message returns the zero PieceBlock.
func (pc *PeerConnection) HandleMessage(msg *Message) (*PieceBlock, error) {
	switch msg.ID {
	case 0:
		pc.Choked = true
		return nil, nil
	case 1:
		pc.Choked = false
		return nil, nil
	case 4:
		// index := binary.BigEndian.Uint32(msg.Payload)
		// pc.setPiece(index)
		return nil, nil

	case 5:
		pc.Bitfield = msg.Payload
		return nil, nil
	case 7:
		return ParsePiece(msg.Payload), nil
	}
	return nil, nil
}

func (pc *PeerConnection) setPiece(index uint32) {
	panic("unimplemented")
}

func (pc *PeerConnection) ReadMessage() (*Message, error) {
	lengthBytes := make([]byte, 4)

	var length uint32
	for {
		_, err := io.ReadFull(pc.Conn, lengthBytes)
		if err != nil {
			return nil, err
		}

		length = binary.BigEndian.Uint32(lengthBytes)

		if length != 0 {
			break
		}
	}

	if length > 17000 {
		return nil, errors.New("malformed bytes")
	}

	payload := make([]byte, length)

	_, err := io.ReadFull(pc.Conn, payload)
	if err != nil {
		return nil, err
	}

	return &Message{ID: payload[0], Payload: payload[1:]}, nil
}

func Connect(peer Peer, infohash [20]byte, peerID [20]byte) (PeerConnection, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port))), 2*time.Second)
	if err != nil {
		return PeerConnection{}, err
	}

	err = conn.SetDeadline(time.Now().Add(6 * time.Second))
	if err != nil {
		return PeerConnection{}, err
	}
	h := Handshake{infohash, peerID}

	bytes, err := h.Serialize()
	if err != nil {
		return PeerConnection{}, err
	}
	_, err = conn.Write(bytes[:])
	if err != nil {
		return PeerConnection{}, err
	}
	buf := make([]byte, 68)

	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return PeerConnection{}, err
	}

	parsed, err := ParseHandshake(buf)
	if err != nil {
		_ = conn.Close()
		return PeerConnection{}, err
	}

	if parsed.InfoHash != infohash {
		_ = conn.Close()
		return PeerConnection{}, errors.New("info hash differs from the original that was sent, closing connection")
	}

	connection := PeerConnection{
		Conn:     conn,
		Peer:     peer,
		InfoHash: infohash,
		PeerID:   peerID,
	}

	return connection, nil
}
