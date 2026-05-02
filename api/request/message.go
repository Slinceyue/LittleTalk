package request

import "context"

type TalkMessageRequest struct {
	FromID      uint   `json:"from_id"`
	ToID        uint   `json:"to_id"`
	RoomID      uint   `json:"room_id"`
	MessageType int8   `json:"message_type"`
	Content     string `json:"content"`
	FileID      uint   `json:"file_id"`
}
type FriendMessageRequest struct {
	FromID uint
	ToID   uint
}
type MessageContext struct {
	MsgType string
	FriendMessageRequest
	TalkMessageRequest
	Ctx context.Context
}
