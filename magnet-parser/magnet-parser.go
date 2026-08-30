package parser

import (
	"fmt"
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

	return magnetURI, nil
}

func mapKeysToMagnetURI(values url.Values) MagnetURI {
	var magnetURI MagnetURI
	for key, values := range values {
		fmt.Printf("%s: %s\n", key, strings.Join(values, ", "))
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
