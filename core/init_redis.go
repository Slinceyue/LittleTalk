package core

import (
	"LittleTalk/global"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func InitRedis() *redis.Client {
	addr := fmt.Sprintf("%s:%d", global.Config.Redis.Host, global.Config.Redis.Port)
	log.Printf("[Redis] 正在连接 %s ...", addr)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: global.Config.Redis.Password,
		DB:       global.Config.Redis.DB,
		PoolSize: global.Config.Redis.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("[Redis] 连接失败: %v", err)
		log.Printf("[Redis] 请确保 Redis 服务已启动，且密码正确")
	} else {
		log.Printf("[Redis] 连接成功: %s", res)
	}

	return rdb
}
