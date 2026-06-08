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

// TO DO: WebSeed can also be a slice. It can have multiple values. Update this and then the parser and the tests.
var KeysArr = []string{ExactTopic, DisplayName, ExactLength, AddressTracker, WebSeed, AcceptableSource, ExactSource, KeywordTopic, ManifestTopic, SelectOnly, Peer}
