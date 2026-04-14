package models

import "LittleTalk/models/enum"

type Message struct {
	Model
	FromUser    User             `gorm:"foreignKey:FromID;"`
	FromID      uint             `gorm:"not null;index" json:"from_id"`
	ToUser      User             `gorm:"foreignKey:ToID;"`
	ToID        uint             `gorm:"index" json:"to_id"`
	Room        Room             `gorm:"foreignKey:RoomID"`
	RoomID      uint             `gorm:"index" json:"room_id"`
	MessageType enum.MessageType `gorm:"default:file" json:"message_type"`
	IsRead      bool             `gorm:"default:false" json:"is_read"` // 是否已读

	Content string `gorm:"size:1024;not null"json:"content"`

	FileID uint `gorm:"index" json:"file_id"`
	File   File `gorm:"foreignKey:FileID;"`
}
