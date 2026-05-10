package request

// CreateRoomRequest 创建群聊请求
type CreateRoomRequest struct {
	Name string `json:"name" binding:"required"`
}

// JoinRoomRequest 加入群聊请求
type JoinRoomRequest struct {
	RoomID uint `json:"room_id" binding:"required"`
}

// UpdateRoomRequest 更新群信息请求
type UpdateRoomRequest struct {
	RoomID  uint   `json:"room_id" binding:"required"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Intro   string `json:"intro"`
}

// SetAdminRequest 设置管理员请求
type SetAdminRequest struct {
	RoomID      uint `json:"room_id" binding:"required"`
	TargetUserID uint `json:"target_user_id" binding:"required"`
	IsAdmin     bool `json:"is_admin"`
}

// KickMemberRequest 踢出成员请求
type KickMemberRequest struct {
	RoomID       uint `json:"room_id" binding:"required"`
	TargetUserID uint `json:"target_user_id" binding:"required"`
}

// TransferOwnerRequest 转让群主请求
type TransferOwnerRequest struct {
	RoomID       uint `json:"room_id" binding:"required"`
	TargetUserID uint `json:"target_user_id" binding:"required"`
}

// GroupMessageRequest 发送群消息请求
type GroupMessageRequest struct {
	RoomID      uint   `json:"room_id" binding:"required"`
	Content     string `json:"content" binding:"required"`
	MessageType int    `json:"message_type"`
}

// SearchRoomRequest 搜索群聊请求
type SearchRoomRequest struct {
	Keyword string `json:"keyword" binding:"required"`
}

// InviteMembersRequest 邀请成员入群请求
type InviteMembersRequest struct {
	RoomID       uint   `json:"room_id" binding:"required"`
	TargetUserIDs []uint `json:"target_user_ids" binding:"required"`
}
