package models

type FriendRequest struct {
	Model
	FromUser   User `gorm:"foreignKey:FromUserID"`
	FromUserID uint `gorm:"not null;index" json:"from_user_id"` // 申请人
	ToUser     User `gorm:"foreignKey:ToUserID"`
	ToUserID   uint `gorm:"not null;index" json:"to_user_id"` // 被申请人
	Status     int8 `gorm:"default:0" json:"status"`
}
