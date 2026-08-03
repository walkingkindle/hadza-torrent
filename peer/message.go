package peer

type Message struct {
	ID      uint8
	Payload []byte
}

// const (
// 	MsgChoke         messageID = 0
// 	MsgUnchoke       messageID = 1
// 	MsgInterested    messageID = 2
// 	MsgNotInterested messageID = 3
// 	MsgHave          messageID = 4
// 	MsgBitfield      messageID = 5
// 	MsgRequest       messageID = 6
// 	MsgPiece         messageID = 7
// 	MsgCancel        messageID = 8
// )
