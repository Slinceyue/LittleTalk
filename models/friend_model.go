package models

import "LittleTalk/models/enum"

type Friend struct {
	UserID      uint              `gorm:"primaryKey;uniqueIndex:idx_user_friend" json:"user_id"`
	FriendID    uint              `gorm:"primaryKey;uniqueIndex:idx_user_friend" json:"friend_id"`
	Remark      string            `gorm:"size:32" json:"remark"`
	Status      enum.FriendStatus `gorm:"default:1" json:"status"`
	UserModel   User              `gorm:"foreignKey:UserID;references:ID" json:"-"`
	FriendModel User              `gorm:"foreignKey:FriendID;references:ID" json:"-"`
}
