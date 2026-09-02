package tracker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	bencodeparser "torrent-client-go/bencode-decoder"
	"torrent-client-go/peer"
)

type HTTPTrackerClient struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

func (c *HTTPTrackerClient) Announce(ctx context.Context, request peer.AnnounceRequest) (*peer.TrackersResponse, error) {
	u := c.BaseURL

	params := url.Values{
		"info_hash": []string{string(request.InfoHash[:])},
		"peer_id":   []string{string(request.PeerID[:])},
		"port":      []string{strconv.Itoa(6881)},
	}
	if request.State != nil {
		state := request.State()

		params.Set("uploaded", strconv.FormatInt(state.Uploaded, 10))
		params.Set("downloaded", strconv.FormatInt(state.Downloaded, 10))
		params.Set("compact", "1")
		params.Set("left", strconv.FormatInt(state.Left, 10))

		if state.Downloaded == 0 {
			params.Add("event", "started")
		}
	}

	u.RawQuery = params.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v")
	}

	resp, err := c.HTTPClient.Do(httpReq)

	responseByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to convert response into the response body")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from peer request %v", resp.StatusCode)
	}

	result, err := bencodeparser.Decode(responseByte)
	if err != nil {
		return nil, fmt.Errorf("unexpected parsing error, %s", err)
	}

	return parseBodyIntoPeerStruct(result)
}
