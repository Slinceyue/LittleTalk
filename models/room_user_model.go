package models

type RoomUser struct {
	Room   Room `gorm:"foreignKey:RoomID"`
	RoomID uint `gorm:"primaryKey;uniqueIndex:idx_room_user" json:"room_id"`
	User   User `gorm:"foreignKey:UserID"`
	UserID uint `gorm:"primaryKey;uniqueIndex:idx_room_user" json:"user_id"`
}
