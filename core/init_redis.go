package core

import (
	"LittleTalk/global"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

func InitRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", global.Config.Redis.Host, global.Config.Redis.Port),
		Password: global.Config.Redis.Password, // no password set
		DB:       global.Config.Redis.DB,
		PoolSize: global.Config.Redis.PoolSize,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	res, err := rdb.Ping(ctx).Result()
	if err != nil {
		println("redis连接失败")
	} else {
		println("redis连接成功", res)
	}
	return rdb
}
