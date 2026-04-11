package models

type Message struct {
	Model
	FromID  uint   `gorm:"not null;index" json:"from_id"`
	ToID    uint   `gorm:"not null;index" json:"to_id"`
	Content string `gorm:"size:1024;not null"json:"content"`
	IsRead  bool   `gorm:"default:false" json:"is_read"` // 是否已读
}
