package parser

type MagnetURI struct {
	ExactTopic    string
	DisplayName   string
	Length        string
	ExactLength   string
	ExactSource   string
	KeywordTopic  string
	ManifestTopic string
	WebSeed       string
	AcceptSource  string
	SelectOnly    string
	Infohash      string
	Name          string
	Trackers      []string
	Peer          string
}

var KeysArr = []string{ExactTopic, DisplayName, ExactLength, AddressTracker, WebSeed, AcceptableSource, Peer}
