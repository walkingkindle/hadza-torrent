package tracker

import (
	"context"
	"net/url"

	"torrent-client-go/peer"

	"github.com/gorilla/websocket"
)

type WSSTrackerClient struct {
	BaseURL *url.URL
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

	message := buildAnnounceMessage(request)

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

func buildAnnounceMessage(request peer.AnnounceRequest) any {
	panic("unimplemented")
}
