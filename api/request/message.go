package request

import "context"

type TalkMessageRequest struct {
	Type        string `json:"type"`
	FromID      uint   `json:"from_id"`
	ToID        uint   `json:"to_id"`
	RoomID      uint   `json:"room_id"`
	MessageType int8   `json:"message_type"`
	Content     string `json:"content"`
	FileID      uint   `json:"file_id"`
	MsgID       string `json:"msg_id,omitempty"`
	FileURL     string `json:"file_url,omitempty"`
	FileName    string `json:"file_name,omitempty"`
}
type FriendMessageRequest struct {
	FromID       uint
	ToID         uint
	FromUsername string
}
type MessageContext struct {
	MsgType string
	FriendMessageRequest
	TalkMessageRequest
	Ctx context.Context
}
