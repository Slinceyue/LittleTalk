package cache

import (
	"LittleTalk/global"
	"context"
	"fmt"
	"time"
)

func SetUserLoginStatus(ctx context.Context, userID uint, token string) error {
	err := global.RDB.Set(ctx, fmt.Sprintf("user:token:%d", userID), token, ExpireToken()).Err()
	return err
}
func SetUserUnLoginStatus(ctx context.Context, userID uint) error {
	err := global.RDB.Del(ctx, fmt.Sprintf("user:token:%d", userID)).Err()
	return err
}
func SetUserOnlineStatus(ctx context.Context, userID uint) error {
	err := global.RDB.Set(ctx, fmt.Sprintf("user:online:%d", userID), true, ExpireStatus()).Err()
	return err
}
func SetUserOfflineStatus(ctx context.Context, userID uint) error {
	err := global.RDB.Del(ctx, fmt.Sprintf("user:online:%d", userID)).Err()
	return err
}
func GetUserOnlineStatus(ctx context.Context, userID uint) bool {
	key := fmt.Sprintf("user:online:%d", userID)
	// Bool() 会忽略错误：key 不存在 → false；存在 → true
	exists, _ := global.RDB.Get(ctx, key).Bool()
	return exists
}
func UserHeartBeat(ctx context.Context, userID uint) error {
	expire := time.Duration(global.Config.Redis.OnlineExpire) * time.Second
	return global.RDB.Expire(ctx, fmt.Sprintf("user:online:%d", userID), expire).Err()
}
func GetUserListOnlineStatus(ctx context.Context, userIDs []uint) []bool {
	status := make([]bool, len(userIDs)) // 预分配
	for i, userID := range userIDs {
		status[i] = GetUserOnlineStatus(ctx, userID)
	}
	return status
}
