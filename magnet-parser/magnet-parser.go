package parser

import (
	"fmt"
	"strings"
)

const MAGNETSTART = "magnet:?"

func ParseMagnet(magnetLink string) (MagnetURI, error) {
	if !isAMagnet(magnetLink) {
		return MagnetURI{}, ErrInvalidMagnetLink
	}

	_, cutMagnet, found := strings.Cut(magnetLink, MAGNETSTART)

	if !found {
		return MagnetURI{}, ErrInvalidMagnetLink
	}

	magnetURI := mapKeysToMagnetURI(extractKeysFromLink(strings.Split(cutMagnet, "&")))

	if magnetURI.ExactTopic == "" {
		return MagnetURI{}, ErrMissingExactTopic
	}

	return magnetURI, nil
}

func extractKeysFromLink(parts []string) map[string][]string {
	KeysMap := map[string][]string{}
	for i := range parts {
		for j := range KeysArr {
			if strings.HasPrefix(parts[i], KeysArr[j]) {
				_, value, found := strings.Cut(parts[i], fmt.Sprintf("%s=", KeysArr[j]))

				if found {
					KeysMap[KeysArr[j]] = append(KeysMap[KeysArr[j]], value)
				}
			}
		}
	}
	return KeysMap
}

func mapKeysToMagnetURI(KeysMap map[string][]string) MagnetURI {
	var magnetURI MagnetURI
	for p := range KeysMap {
		switch p {
		case ExactTopic:
			magnetURI.ExactTopic = KeysMap[p][0]
		case DisplayName:
			magnetURI.DisplayName = KeysMap[p][0]
		case ExactLength:
			magnetURI.ExactLength = KeysMap[p][0]
		case AddressTracker:
			magnetURI.Trackers = append(magnetURI.Trackers, KeysMap[p]...)
		case WebSeed:
			magnetURI.WebSeed = KeysMap[p][0]

		case AcceptableSource:
			magnetURI.AcceptSource = KeysMap[p][0]

		case ExactSource:
			magnetURI.ExactSource = KeysMap[p][0]

		case KeywordTopic:
			magnetURI.KeywordTopic = KeysMap[p][0]

		case ManifestTopic:
			magnetURI.ManifestTopic = KeysMap[p][0]

		case SelectOnly:
			magnetURI.SelectOnly = KeysMap[p][0]
		case Peer:
			magnetURI.Peer = KeysMap[p][0]

		}
	}
	return magnetURI
}

func isAMagnet(magnet string) bool {
	return strings.HasPrefix(magnet, MAGNETSTART)
}
