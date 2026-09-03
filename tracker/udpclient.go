package tracker

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/url"
	"time"

	"torrent-client-go/peer"
)

type UDPTrackerClient struct {
	BaseURL *url.URL
}

const protocolID uint64 = 0x41727101980
const (
	actionConnect  uint32 = 0
	actionAnnounce uint32 = 1
)

func (c *UDPTrackerClient) Announce(
	ctx context.Context,
	request peer.AnnounceRequest,
) (*peer.TrackersResponse, error) {
	addr, err := net.ResolveUDPAddr("udp", c.BaseURL.Host)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	if err := setDeadline(ctx, conn); err != nil {
		return nil, err
	}

	connID, err := c.connect(ctx, conn)
	if err != nil {
		return nil, err
	}

	transactionID := rand.Uint32()

	packet := buildAnnouncePacket(connID, transactionID, request)

	_, err = conn.Write(packet)
	if err != nil {
		return nil, err
	}

	return readAnnounceResponse(conn, transactionID)
}

func setDeadline(ctx context.Context, conn *net.UDPConn) error {
	if deadline, ok := ctx.Deadline(); ok {
		return conn.SetDeadline(deadline)
	}

	return conn.SetDeadline(time.Now().Add(10 * time.Second))
}

func readAnnounceResponse(
	conn *net.UDPConn,
	transactionID uint32,
) (*peer.TrackersResponse, error) {
	buf := make([]byte, 2048)

	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	if n < 20 {
		return nil, errors.New("announce response too short")
	}

	response := buf[:n]

	action := binary.BigEndian.Uint32(response[0:4])
	responseTransactionID := binary.BigEndian.Uint32(response[4:8])

	if action == 3 {
		message := string(response[8:])
		return nil, fmt.Errorf("tracker error: %s", message)
	}

	if action != actionAnnounce {
		return nil, fmt.Errorf("unexpected action: %d", action)
	}

	if responseTransactionID != transactionID {
		return nil, errors.New("transaction ID mismatch")
	}

	interval := binary.BigEndian.Uint32(response[8:12])
	leechers := binary.BigEndian.Uint32(response[12:16])
	seeders := binary.BigEndian.Uint32(response[16:20])

	peers := make([]peer.Peer, 0)

	for offset := 20; offset+6 <= n; offset += 6 {
		ip := net.IPv4(
			response[offset],
			response[offset+1],
			response[offset+2],
			response[offset+3],
		)

		port := binary.BigEndian.Uint16(
			response[offset+4 : offset+6],
		)

		peers = append(peers, peer.Peer{
			IP:   ip,
			Port: port,
		})
	}

	return &peer.TrackersResponse{
		Peers:    peers,
		Interval: int(interval),
		Seeders:  int(seeders),
		Leechers: int(leechers),
	}, nil
}

func (c *UDPTrackerClient) connect(
	ctx context.Context,
	conn *net.UDPConn,
) (uint64, error) {
	transactionID := rand.Uint32()

	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[0:8], protocolID)
	binary.BigEndian.PutUint32(buf[8:12], actionConnect)
	binary.BigEndian.PutUint32(buf[12:16], transactionID)

	_, err := conn.Write(buf)
	if err != nil {
		return 0, err
	}

	response := make([]byte, 16)

	if _, err := io.ReadFull(conn, response); err != nil {
		return 0, nil
	}

	action := binary.BigEndian.Uint32(response[0:4])
	responseTransactionID := binary.BigEndian.Uint32(response[4:8])
	connectionID := binary.BigEndian.Uint64(response[8:16])

	if action != actionConnect {
		return 0, fmt.Errorf("unexpected action: %d", action)
	}

	if responseTransactionID != transactionID {
		return 0, errors.New("transaction ID mismatch")
	}

	return connectionID, nil
}

func buildAnnouncePacket(
	connectionID uint64,
	transactionID uint32,
	request peer.AnnounceRequest,
) []byte {
	buf := make([]byte, 98)

	offset := 0

	binary.BigEndian.PutUint64(buf[offset:], connectionID)
	offset += 8

	binary.BigEndian.PutUint32(buf[offset:], actionAnnounce)
	offset += 4

	binary.BigEndian.PutUint32(buf[offset:], transactionID)
	offset += 4

	copy(buf[offset:], request.InfoHash[:])
	offset += 20

	copy(buf[offset:], request.PeerID[:])
	offset += 20

	var state peer.DownloadState

	if request.State != nil {
		state = request.State()
	}

	// downloaded
	binary.BigEndian.PutUint64(buf[offset:], uint64(state.Downloaded))
	offset += 8

	// left
	binary.BigEndian.PutUint64(buf[offset:], uint64(state.Left))
	offset += 8

	// uploaded
	binary.BigEndian.PutUint64(buf[offset:], uint64(state.Uploaded))
	offset += 8

	// event
	binary.BigEndian.PutUint32(buf[offset:], 0)
	offset += 4

	// IP address = 0
	binary.BigEndian.PutUint32(buf[offset:], 0)
	offset += 4

	// key
	binary.BigEndian.PutUint32(buf[offset:], rand.Uint32())
	offset += 4

	// num_want
	binary.BigEndian.PutUint32(buf[offset:], ^uint32(0))
	offset += 4

	// port
	binary.BigEndian.PutUint16(buf[offset:], 6881)

	return buf
}
