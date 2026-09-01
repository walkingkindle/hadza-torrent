package tracker

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	bencodeparser "torrent-client-go/bencode-decoder"
	"torrent-client-go/peer"
)

func getResponse(url string) (peer.TrackersResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		return peer.TrackersResponse{}, fmt.Errorf("failed to send request: %v", err)
	}

	responseByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return peer.TrackersResponse{}, errors.New("failed to convert response into the response body")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return peer.TrackersResponse{}, fmt.Errorf("unexpected status code from peer request %v", resp.StatusCode)
	}

	result, err := bencodeparser.Decode(responseByte)
	if err != nil {
		return peer.TrackersResponse{}, fmt.Errorf("unexpected parsing error, %s", err)
	}

	return parseBodyIntoPeerStruct(result)
}

func buildTrackerURL(request peer.AnnounceRequest, trackerURL string) (string, error) {
	base, err := url.Parse(trackerURL)
	if err != nil {
		return "", err
	}

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

	base.RawQuery = params.Encode()

	return base.String(), nil
}
