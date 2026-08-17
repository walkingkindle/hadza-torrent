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

func buildTrackerURL(infoHash []byte, torrentlength int, announce string, peerId string, uploaded string, downloaded string) (string, error) {
	base, err := url.Parse(announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(infoHash)},
		"peer_id":    []string{peerId},
		"port":       []string{strconv.Itoa(6881)},
		"uploaded":   []string{uploaded},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(torrentlength)},
	}

	base.RawQuery = params.Encode()

	return base.String(), nil
}
