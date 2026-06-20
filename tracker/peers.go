// Package tracker is responsible for creating the connection to trackers.
package tracker

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"

	bencodeparser "torrent-client-go/bencode-decoder"
	"torrent-client-go/types"
)

func GetPeers(torrent types.TorrentFile, peerId string) ([]types.Peer, error) {
	url, err := buildTrackerURL(torrent.InfoHash[:],
		torrent.Length, torrent.Announce, peerId)
	if err != nil {
		return []types.Peer{}, errors.New("failed parsing url")
	}

	return getResponse(url)
}

func getResponse(url string) ([]types.Peer, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []types.Peer{}, fmt.Errorf("failed to send request: %v", err)
	}

	responseByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return []types.Peer{}, errors.New("failed to convert response into the response body")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []types.Peer{}, fmt.Errorf("unexpected status code from peer request %v", resp.StatusCode)
	}

	result, _, err := bencodeparser.Decode(responseByte)
	if err != nil {
		return []types.Peer{}, fmt.Errorf("unexpected parsing error, %s", err)
	}

	return parseBodyIntoPeerStruct(result)
}

func parseBodyIntoPeerStruct(result map[string]any) ([]types.Peer, error) {
	peers := []types.Peer{}

	for range result {
		peer := types.Peer{}
		ip, ok := result["ip"].(net.IP)

		if !ok {
			return []types.Peer{}, errors.New("could not parse, ip not found")
		}
		peer.IP = ip

		port, ok := result["port"].(uint16)

		if !ok {
			return []types.Peer{}, errors.New("could not parse port")
		}
		peer.Port = port

		peers = append(peers, peer)
	}

	return peers, nil
}

func buildTrackerURL(infoHash []byte, torrentlength int, announce string, peerId string) (string, error) {
	base, err := url.Parse(announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(infoHash)},
		"peer_id":    []string{peerId},
		"port":       []string{strconv.Itoa(6881)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(torrentlength)},
	}

	base.RawQuery = params.Encode()

	return base.String(), nil
}
