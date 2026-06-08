package parser

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStringReceivedIsMagnet(t *testing.T) {
	magnetLink := "magnet:?xt=urn:btih:c12fe1c06bba254a9dc9f519b335aa7c1367a88a&dn=My+File&tr=udp%3A%2F%2Ftracker.example.com%3A6969"

	result := isAMagnet(magnetLink)

	if !result {
		t.Errorf("Magnet expected to be true for the string valid magnet, returned %t", result)
	}
}

func TestStringReceivedIsNotMagnet(t *testing.T) {
	magnetLink := "whatever"

	result := isAMagnet(magnetLink)

	if result {
		t.Errorf("Expected false but returned %t", result)
	}
}

func TestParseMagnet(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        MagnetURI
		wantErr     bool
		expectedErr error
	}{
		{
			name:  "Valid Ubuntu Magnet (Full Base Case)",
			input: "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82&dn=ubuntu-22.04.3-desktop-amd64.iso&xl=4718592000&tr=udp%3A%2F%://openbittorrent.com%3A80&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&ws=http%3A%2F%://ubuntu.com%2F22.04%2Fubuntu-22.04.3-desktop-amd64.iso&as=http%3A%2F%://example.com%2Fubuntu.iso&x.pe=192.0.2.1:6881",
			want: MagnetURI{
				ExactTopic:   "urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
				DisplayName:  "ubuntu-22.04.3-desktop-amd64.iso",
				ExactLength:  "4718592000",
				Trackers:     []string{"udp%3A%2F%://openbittorrent.com%3A80", "udp%3A%2F%2Ftracker.opentrackr.org%3A1337"},
				WebSeed:      "http%3A%2F%://ubuntu.com%2F22.04%2Fubuntu-22.04.3-desktop-amd64.iso",
				AcceptSource: "http%3A%2F%://example.com%2Fubuntu.iso",
				Peer:         "192.0.2.1:6881",
			},
			wantErr: false,
		},
		{
			name:  "Minimal Magnet Link (Topic Only)",
			input: "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
			want: MagnetURI{
				ExactTopic: "urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
			},
			wantErr: false,
		},
		{
			name:  "Base32 Infohash (Older v1 standard)",
			input: "magnet:?xt=urn:btih:MR6OCMREDC5K4HUT34Y5ORSTC3LUM7MC&dn=Documentary",
			want: MagnetURI{
				ExactTopic:  "urn:btih:MR6OCMREDC5K4HUT34Y5ORSTC3LUM7MC",
				DisplayName: "Documentary",
			},
			wantErr: false,
		},
		{
			name:  "BitTorrent v2 Infohash (SHA-256 / 64 characters)",
			input: "magnet:?xt=urn:btmh:1220cafac75a2cb703e226a0efbf23004f2cf77b94902120e2ef59842a66e64edc23&dn=SampleVideo",
			want: MagnetURI{
				ExactTopic:  "urn:btmh:1220cafac75a2cb703e226a0efbf23004f2cf77b94902120e2ef59842a66e64edc23",
				DisplayName: "SampleVideo",
			},
			wantErr: false,
		},
		{
			name:  "Multiple Trackers and WebSeeds (Slice appending test)",
			input: "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82&tr=udp%3A%2F%2Ftracker1.com&tr=udp%3A%2F%2Ftracker2.com&ws=http%3A%2F%2Fseed1.com&ws=http%3A%2F%2Fseed2.com",
			want: MagnetURI{
				ExactTopic: "urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
				Trackers:   []string{"udp%3A%2F%2Ftracker1.com", "udp%3A%2F%2Ftracker2.com"},
				// Note: If your struct holds multiple webseeds, adjust this to a slice:
				WebSeed: "http%3A%2F%2Fseed1.com",
			},
			wantErr: false,
		},
		{
			name:  "Out of Order Parameters",
			input: "magnet:?dn=Movie.mp4&tr=udp%3A%2F%2Ftracker.org&xt=urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
			want: MagnetURI{
				ExactTopic:  "urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
				DisplayName: "Movie.mp4",
				Trackers:    []string{"udp%3A%2F%2Ftracker.org"},
			},
			wantErr: false,
		},
		{
			name:  "Display Name with Spaces and Special Characters",
			input: "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82&dn=Ubuntu%20Linux%2022.04%20(LTS)",
			want: MagnetURI{
				ExactTopic:  "urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
				DisplayName: "Ubuntu%20Linux%2022.04%20(LTS)", // Change if your parser decodes %20 to spaces
			},
			wantErr: false,
		},
		{
			name:  "Empty Values For Optional Fields",
			input: "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82&dn=&xl=",
			want: MagnetURI{
				ExactTopic:  "urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
				DisplayName: "",
				ExactLength: "",
			},
			wantErr: false,
		},
		{
			name: "Exact Source, Keyword Topic, Manifest Topic and Select Only",
			input: "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82" +
				"&xs=http%3A%2F%2Fcache.example.com%2Ffile" +
				"&kt=ubuntu+linux+iso" +
				"&mt=http%3A%2F%2Fexample.com%2Fmeta.txt" +
				"&so=0-2",
			want: MagnetURI{
				ExactTopic:    "urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
				ExactSource:   "http%3A%2F%2Fcache.example.com%2Ffile",
				KeywordTopic:  "ubuntu+linux+iso",
				ManifestTopic: "http%3A%2F%2Fexample.com%2Fmeta.txt",
				SelectOnly:    "0-2",
			},
			wantErr: false,
		},
		{
			name:  "Unknown Keys Are Ignored",
			input: "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82&foo=bar&unknown=123",
			want: MagnetURI{
				ExactTopic: "urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
			},
			wantErr: false,
		},
		{
			name:  "Key Without Value Separator Is Skipped",
			input: "magnet:?xt=urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82&dn&tr",
			want: MagnetURI{
				ExactTopic: "urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
			},
			wantErr: false,
		},
		{
			name:        "Error: Empty Input String",
			input:       "",
			want:        MagnetURI{},
			wantErr:     true,
			expectedErr: ErrInvalidMagnetLink,
		},
		{
			name:        "Error: Missing Magnet Scheme prefix",
			input:       "?xt=urn:btih:08ada5a7a6183aae1e09d831df674416d7467d82",
			want:        MagnetURI{},
			wantErr:     true,
			expectedErr: ErrInvalidMagnetLink,
		},
		{
			name:        "Error: HTTP Link Passed Instead",
			input:       "https://ubuntu.com",
			want:        MagnetURI{},
			wantErr:     true,
			expectedErr: ErrInvalidMagnetLink,
		},
		{
			name:        "Error: Missing Exact Topic (xt) Parameter",
			input:       "magnet:?dn=ubuntu-22.04.iso&xl=4718592000",
			want:        MagnetURI{},
			wantErr:     true, // Standard magnets usually require an 'xt' topic to be valid
			expectedErr: ErrMissingExactTopic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMagnet(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseMagnet() expected error %v, got nil", tt.expectedErr)
				}
				if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Fatalf("ParseMagnet() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseMagnet() unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseMagnet() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
