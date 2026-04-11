package models

import "LittleTalk/models/enum"

type Friend struct {
	User
	UserID   uint              `gorm:"primary_key;index:idx_user_friend;uniqueIndex" json:"user_id"`
	FriendID uint              `gorm:"primary_key;index:idx_user_friend;uniqueIndex" json:"friend_id"`
	Remark   string            `gorm:"size:32" json:"remark"` // 好友备注
	Status   enum.FriendStatus `gorm:"default:1" json:"status"`
}
