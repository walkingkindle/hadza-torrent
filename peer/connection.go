// Package peer tries to handshake with the return peers and the connection struct
package peer

import (
	"errors"
	"io"
	"net"
	"strconv"
	"time"
)

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
