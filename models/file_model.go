package models

type File struct {
	Model
	UserID    uint   `gorm:"not null;index"`
	User      User   `gorm:"foreignKey:UserID"`
	FileType  string `gorm:"not null;index;comment:'文件类型'"`
	RealName  string `gorm:"not null;comment:'原始文件名'"` // 用户看到的
	SavedName string `gorm:"not null;comment:'存储文件名'"` // UUID 名称
	Src       string `gorm:"not null;comment:'访问路径'"`
	Size      int64  `gorm:"not null;comment:'文件大小'"`
}
