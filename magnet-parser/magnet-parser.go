package parser

import (
	"errors"
	"net/url"
	"strings"
)

const MAGNETSTART = "magnet:?"

func ParseMagnet(magnetLink string) (MagnetURI, error) {
	if magnetLink == "" || !IsAMagnet(magnetLink) {
		return MagnetURI{}, ErrInvalidMagnetLink
	}

	u, err := url.Parse(magnetLink)
	if err != nil {
		return MagnetURI{}, err
	}

	magnetURI := mapKeysToMagnetURI(u.Query())

	if magnetURI.ExactTopic == "" {
		return MagnetURI{}, ErrMissingExactTopic
	}
	infoHash, ok := magnetURI.GetInfoHash()

	if !ok {
		return MagnetURI{}, errors.New("Invalid infohash value")
	}
	magnetURI.Infohash = infoHash
	return magnetURI, nil
}

func mapKeysToMagnetURI(values url.Values) MagnetURI {
	var magnetURI MagnetURI
	for key, values := range values {
		switch key {
		case ExactTopic:
			magnetURI.ExactTopic = values[0]
		case DisplayName:
			magnetURI.DisplayName = values[0]
		case ExactLength:
			magnetURI.ExactLength = values[0]
		case AddressTracker:
			magnetURI.Trackers = append(magnetURI.Trackers, values...)
		case WebSeed:
			magnetURI.WebSeed = values[0]

		case AcceptableSource:
			magnetURI.AcceptSource = values[0]

		case ExactSource:
			magnetURI.ExactSource = values[0]

		case KeywordTopic:
			magnetURI.KeywordTopic = values[0]

		case ManifestTopic:
			magnetURI.ManifestTopic = values[0]

		case SelectOnly:
			magnetURI.SelectOnly = values[0]
		case Peer:
			magnetURI.Peer = values[0]

		}
	}

	return magnetURI
}

func IsAMagnet(magnet string) bool {
	return strings.HasPrefix(magnet, MAGNETSTART)
}

func (m MagnetURI) GetInfoHash() (string, bool) {
	BTIHPrefix := "urn:btih:"
	if strings.HasPrefix(m.ExactTopic, BTIHPrefix) {
		return strings.TrimPrefix(m.ExactTopic, BTIHPrefix), true
	}
	return "", false // Not a BitTorrent magnet link or missing prefix
}
