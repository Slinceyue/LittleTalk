package global

import (
	"LittleTalk/conf"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var (
	Config *conf.Config
	DB     *gorm.DB
	RDB    *redis.Client
)
