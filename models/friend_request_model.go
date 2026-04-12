package models

import "LittleTalk/models/enum"

type FriendRequest struct {
	Model
	FromUserID uint                     `gorm:"not null;index" json:"from_user_id"` // 申请人
	ToUserID   uint                     `gorm:"not null;index" json:"to_user_id"`   // 被申请人
	Status     enum.FriendRequestStatus `gorm:"default:0" json:"status"`
}
