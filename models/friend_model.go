package models

type Friend struct {
	UserID      uint   `gorm:"primaryKey;index:idx_user_friend"` // 去掉 unique
	FriendID    uint   `gorm:"primaryKey;index:idx_user_friend"` // 去掉 unique
	Remark      string `gorm:"size:32" json:"remark"`
	Status      int8   `gorm:"default:1" json:"status"`
	UserModel   User   `gorm:"foreignKey:UserID;references:ID" json:"-"`
	FriendModel User   `gorm:"foreignKey:FriendID;references:ID" json:"-"`
}
