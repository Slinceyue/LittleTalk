package models

// RoomUser 群成员关系
type RoomUser struct {
	Room     Room `gorm:"foreignKey:RoomID"`
	RoomID   uint `gorm:"primaryKey;uniqueIndex:idx_room_user" json:"room_id"`
	User     User `gorm:"foreignKey:UserID"`
	UserID   uint `gorm:"primaryKey;uniqueIndex:idx_room_user" json:"user_id"`
	Role     int8 `gorm:"default:2" json:"role"` // 0-群主 1-管理员 2-成员
	Nickname string `gorm:"size:32;default:''" json:"nickname"` // 群昵称
	JoinTime int64 `gorm:"autoCreateTime" json:"join_time"`
}
