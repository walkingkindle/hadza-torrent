// Package peer tries to handshake with the return peers and the connection struct
package peer

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"time"
)

const extensionProtocolBit byte = 0x10

// addr identifies this peer in log lines.
func (pc *PeerConnection) addr() string {
	return net.JoinHostPort(pc.Peer.IP.String(), strconv.Itoa(int(pc.Peer.Port)))
}

// Send writes a message to the peer and logs it, so that every byte leaving
// the client is visible at debug level.
func (pc *PeerConnection) Send(msg Message) error {
	slog.Debug("-> send", "peer", pc.addr(), "msg", msg.Name(), "payload", len(msg.Payload))

	_, err := pc.Conn.Write(msg.Serialize())

	return err
}

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

func (conn *PeerConnection) WaitForUnchoke() error {
	for conn.Choked {
		msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		_, err = conn.HandleMessage(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// HandleMessage applies msg to the connection state. Only a piece message
// yields a block; every other message returns nil.
func (pc *PeerConnection) HandleMessage(msg *Message) (*PieceBlock, error) {
	switch msg.ID {
	case MsgChoke:
		pc.Choked = true
		slog.Info("peer choked us", "peer", pc.addr())
		return nil, nil
	case MsgUnchoke:
		pc.Choked = false
		slog.Info("peer unchoked us", "peer", pc.addr())
		return nil, nil
	case MsgHave:
		if len(msg.Payload) < 4 {
			return nil, errors.New("malformed have message, want a 4 byte piece index")
		}
		index := binary.BigEndian.Uint32(msg.Payload)
		pc.setPiece(index)
		slog.Info("peer advertised a piece", "peer", pc.addr(), "piece", index, "total_known", pc.PieceCount())
		return nil, nil
	case MsgBitfield:
		pc.Bitfield = msg.Payload
		slog.Info("peer sent bitfield", "peer", pc.addr(), "bytes", len(msg.Payload), "pieces", pc.PieceCount())
		return nil, nil
	case MsgPiece:
		piece, err := ParsePiece(msg.Payload)
		if err != nil {
			return nil, err
		}
		return piece, nil
	}
	return nil, nil
}

// setPiece marks piece i as available on this peer. A peer is free to send
// have messages without ever having sent a bitfield, so grow the bitfield to
// fit rather than assuming one arrived first.
func (pc *PeerConnection) setPiece(i uint32) {
	byteIndex := i / 8

	bitOffset := 7 - (i % 8)

	if int(byteIndex) >= len(pc.Bitfield) {
		grown := make([]byte, byteIndex+1)
		copy(grown, pc.Bitfield)
		pc.Bitfield = grown
	}

	pc.Bitfield[byteIndex] |= 1 << bitOffset
}

// HasPiece reports whether the peer has advertised piece i, via either its
// bitfield or a have message. Requesting a piece a peer does not have is a
// protocol violation and most clients respond by dropping the connection.
func (pc *PeerConnection) HasPiece(i int) bool {
	byteIndex := i / 8

	bitOffset := 7 - (i % 8)

	if i < 0 || byteIndex >= len(pc.Bitfield) {
		return false
	}

	return pc.Bitfield[byteIndex]>>bitOffset&1 == 1
}

// PieceCount is how many pieces the peer has advertised so far.
func (pc *PeerConnection) PieceCount() int {
	count := 0

	for _, b := range pc.Bitfield {
		for bit := range 8 {
			if b>>(7-bit)&1 == 1 {
				count++
			}
		}
	}

	return count
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
		slog.Debug("<- recv", "peer", pc.addr(), "msg", "keep-alive")
	}

	if length > 17000 {
		return nil, errors.New("malformed bytes")
	}

	payload := make([]byte, length)

	_, err := io.ReadFull(pc.Conn, payload)
	if err != nil {
		return nil, err
	}

	msg := &Message{ID: payload[0], Payload: payload[1:]}

	slog.Debug("<- recv", "peer", pc.addr(), "msg", msg.Name(), "payload", len(msg.Payload))

	return msg, nil
}

func Connect(peer Peer, infohash [20]byte, peerID [20]byte) (PeerConnection, error) {
	addr := net.JoinHostPort(peer.IP.String(), strconv.Itoa(int(peer.Port)))

	slog.Debug("dialing peer", "peer", addr)

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		slog.Debug("dial failed", "peer", addr, "err", err)
		return PeerConnection{}, err
	}

	err = conn.SetDeadline(time.Now().Add(6 * time.Second))
	if err != nil {
		return PeerConnection{}, err
	}
	var reserved [8]byte
	reserved[5] |= 0x10
	h := Handshake{reserved, infohash, peerID}

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

	slog.Info("handshake accepted", "peer", addr, "peer_id", fmt.Sprintf("%q", trimPeerID(parsed.PeerID)))

	return PeerConnection{
		Conn:              conn,
		Peer:              peer,
		InfoHash:          infohash,
		PeerID:            peerID,
		Choked:            true,
		SupportsExtension: parsed.SupportsExtensions(),
	}, nil
}

func (conn *PeerConnection) SendInterested() error {
	if err := conn.Send(Message{ID: MsgInterested}); err != nil {
		return err
	}
	return nil
}

// trimPeerID keeps the printable client identifier most peers put at the front
// of their peer id (for example "-qB5000-"), dropping the random tail.
func trimPeerID(id [20]byte) string {
	for i, b := range id {
		if b < 0x20 || b > 0x7e {
			return string(id[:i])
		}
	}

	return string(id[:])
}

func (h Handshake) SupportsExtensions() bool {
	return h.Reserved[5]&extensionProtocolBit != 0
}
