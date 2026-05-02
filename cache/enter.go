package cache

import (
	"LittleTalk/global"
	"time"
)

func ExpireStatus() time.Duration {
	return time.Duration(global.Config.Redis.OnlineExpire) * time.Second
}

func ExpireToken() time.Duration {
	return time.Duration(global.Config.Jwt.Expire) * 24 * time.Hour
}

func DefaultExpire() time.Duration {
	return 1 * time.Hour
}
