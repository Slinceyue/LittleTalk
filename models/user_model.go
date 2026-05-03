package models

type User struct {
	Model
	Username string `gorm:"size:10;unique;not null" json:"username"`
	Password string `gorm:"size:256;not null"        json:"password"`
	BackSalt string `gorm:"size:256" json:"back_salt"`
	Sex      int8   `gorm:"default:0"      json:"sex"`
	Intro    string `gorm:"size:255;default:''" json:"intro"`
	Avatar   string `gorm:"size:255;default:''" json:"avatar"` // 头像路径

	Phone    string `gorm:"size:16;unique;index;default:null" json:"phone"`
	Email    string `gorm:"size:64;unique;index;default:null" json:"email"`
	Birthday string `gorm:"size:20" json:"birthday"`
	Status   int8   `gorm:"default:1" json:"status"` // 1正常 2禁用
	Role     int8   `gorm:"default:1" json:"role"`   // 1普通 2管理员
}
