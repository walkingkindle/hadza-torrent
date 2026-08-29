package peer

import "fmt"

type Message struct {
	ID      uint8
	Payload []byte
}

// Name is the protocol name of the message, for logs.
func (m *Message) Name() string {
	if m == nil {
		return "keep-alive"
	}

	switch m.ID {
	case MsgChoke:
		return "choke"
	case MsgUnchoke:
		return "unchoke"
	case MsgInterested:
		return "interested"
	case MsgNotInterested:
		return "not-interested"
	case MsgHave:
		return "have"
	case MsgBitfield:
		return "bitfield"
	case MsgRequest:
		return "request"
	case MsgPiece:
		return "piece"
	case MsgCancel:
		return "cancel"
	}

	return fmt.Sprintf("unknown(%d)", m.ID)
}

const (
	MsgChoke         uint8 = 0
	MsgUnchoke       uint8 = 1
	MsgInterested    uint8 = 2
	MsgNotInterested uint8 = 3
	MsgHave          uint8 = 4
	MsgBitfield      uint8 = 5
	MsgRequest       uint8 = 6
	MsgPiece         uint8 = 7
	MsgCancel        uint8 = 8
	MsgExtended      uint8 = 20
)
