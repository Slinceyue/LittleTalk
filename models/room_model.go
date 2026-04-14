package models

// Room 房间模型
type Room struct {
	Model
	// Name 房间名称
	Name      string `gorm:"size:255;not null" json:"name"`
	Owner     User   `gorm:"foreignKeyOwnerID"`
	OwnerID   uint   `gorm:"not null;index" json:"owner_id"` // OwnerID 房间所有者ID
	IsPrivate bool   `gorm:"default:false" json:"is_private"`
}
