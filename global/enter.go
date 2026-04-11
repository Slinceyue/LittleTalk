package global

import (
	"LittleTalk/conf"

	"gorm.io/gorm"
)

var (
	Config *conf.Config
	DB     *gorm.DB
)
