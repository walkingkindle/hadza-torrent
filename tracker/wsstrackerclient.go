package tracker

import (
	"context"
	"encoding/json"
	"net/url"

	"torrent-client-go/peer"

	"github.com/gorilla/websocket"
)

type WSSTrackerClient struct {
	BaseURL *url.URL
}
type announceMessage struct {
	Action     string `json:"action"`
	InfoHash   string `json:"info_hash"`
	PeerID     string `json:"peer_id"`
	NumWant    int    `json:"numwant"`
	Uploaded   int64  `json:"uploaded"`
	Downloaded int64  `json:"downloaded"`
	Left       int64  `json:"left"`
	Compact    int    `json:"compact"`
}

func (c *WSSTrackerClient) Announce(
	ctx context.Context,
	request peer.AnnounceRequest,
) (*peer.TrackersResponse, error) {
	dialer := websocket.Dialer{}

	conn, _, err := dialer.DialContext(
		ctx,
		c.BaseURL.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	message, err := buildAnnounceMessage(request)

	if err != nil {
		return nil, err
	}

	if err := conn.WriteMessage(
		websocket.BinaryMessage,
		message,
	); err != nil {
		return nil, err
	}

	_, response, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	return parseTrackerResponse(response)
}

// TODO: Implement these 2 methods lol
func parseTrackerResponse(response []byte) (*peer.TrackersResponse, error) {
	panic("unimplemented")
}

func buildAnnounceMessage(request peer.AnnounceRequest) ([]byte, error) {
	message := announceMessage{
		Action:   "announce",
		InfoHash: string(request.InfoHash[:]),
		PeerID:   string(request.PeerID[:]),
		NumWant:  50,
		Compact:  1,
	}

	if request.State != nil {
		state := request.State()

		message.Uploaded = state.Uploaded
		message.Downloaded = state.Downloaded
		message.Left = state.Left
	}

	return json.Marshal(message)
}
