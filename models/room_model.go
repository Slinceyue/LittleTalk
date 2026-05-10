package models

// Room 房间模型
type Room struct {
	Model
	Name      string `gorm:"size:255;not null" json:"name"`
	Owner     User   `gorm:"foreignKey:OwnerID"`
	OwnerID   uint   `gorm:"not null;index" json:"owner_id"`
	Avatar    string `gorm:"size:255;default:''" json:"avatar"`
	Intro     string `gorm:"size:255;default:''" json:"intro"`
	MemberCnt int    `gorm:"default:0" json:"member_cnt"`
}
