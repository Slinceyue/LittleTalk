package cache

import (
	"LittleTalk/api/response"
	"LittleTalk/global"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// GetUserInfoCache 获取用户信息缓存
func GetUserInfoCache(ctx context.Context, userID uint) (*response.SelfUserResponse, error) {
	key := fmt.Sprintf("user:info:%d", userID)
	data, err := global.RDB.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var userInfo response.SelfUserResponse
	if err := json.Unmarshal([]byte(data), &userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}

// SetUserInfoCache 设置用户信息缓存
func SetUserInfoCache(ctx context.Context, userID uint, userInfo *response.SelfUserResponse) error {
	key := fmt.Sprintf("user:info:%d", userID)
	data, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}

	expire := ExpireUserInfo()
	return global.RDB.Set(ctx, key, data, expire).Err()
}

// DelUserInfoCache 删除用户信息缓存
func DelUserInfoCache(ctx context.Context, userID uint) error {
	key := fmt.Sprintf("user:info:%d", userID)
	return global.RDB.Del(ctx, key).Err()
}

// ExpireUserInfo 获取用户信息缓存过期时间
func ExpireUserInfo() time.Duration {
	expire := global.Config.Cache.UserInfoExpire
	if expire <= 0 {
		expire = 3600 // 默认1小时
	}
	return time.Duration(expire) * time.Second
}

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

// SetUserOnlineStatusPermanent 设置用户在线状态（永不过期）
// WS连接存在时使用，生命周期由WS连接管理，不再依赖心跳过期
func SetUserOnlineStatusPermanent(ctx context.Context, userID uint) error {
	err := global.RDB.Set(ctx, fmt.Sprintf("user:online:%d", userID), true, 0).Err() // 0表示永不过期
	return err
}

func SetUserOfflineStatus(ctx context.Context, userID uint) error {
	err := global.RDB.Del(ctx, fmt.Sprintf("user:online:%d", userID)).Err()
	return err
}

// DelUserOnlineStatus 删除用户在线状态（WS断开时调用）
func DelUserOnlineStatus(ctx context.Context, userID uint) error {
	err := global.RDB.Del(ctx, fmt.Sprintf("user:online:%d", userID)).Err()
	return err
}

func GetUserOnlineStatus(ctx context.Context, userID uint) bool {
	key := fmt.Sprintf("user:online:%d", userID)
	exists, _ := global.RDB.Get(ctx, key).Bool()
	return exists
}

func UserHeartBeat(ctx context.Context, userID uint) error {
	expire := time.Duration(global.Config.Redis.OnlineExpire) * time.Second
	return global.RDB.Expire(ctx, fmt.Sprintf("user:online:%d", userID), expire).Err()
}

// GetUserListOnlineStatus 批量获取在线状态（旧方法，保留兼容性）
func GetUserListOnlineStatus(ctx context.Context, userIDs []uint) []bool {
	status := make([]bool, len(userIDs))
	for i, userID := range userIDs {
		status[i] = GetUserOnlineStatus(ctx, userID)
	}
	return status
}

// GetBatchUserOnlineStatus 批量获取在线状态（Pipeline优化）
func GetBatchUserOnlineStatus(ctx context.Context, userIDs []uint) map[uint]bool {
	if len(userIDs) == 0 {
		return make(map[uint]bool)
	}

	result := make(map[uint]bool, len(userIDs))

	// 使用Pipeline批量查询
	pipe := global.RDB.Pipeline()
	cmds := make(map[uint]*redis.StringCmd, len(userIDs))

	for _, userID := range userIDs {
		key := fmt.Sprintf("user:online:%d", userID)
		cmds[userID] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// 如果Redis出错，返回默认值
		for _, id := range userIDs {
			result[id] = false
		}
		return result
	}

	for userID, cmd := range cmds {
		exists, err := cmd.Bool()
		if err != nil {
			result[userID] = false
		} else {
			result[userID] = exists
		}
	}

	return result
}

// SetBatchUserOnlineStatus 批量设置在线状态
func SetBatchUserOnlineStatus(ctx context.Context, userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}

	pipe := global.RDB.Pipeline()
	expire := ExpireStatus()

	for _, userID := range userIDs {
		key := fmt.Sprintf("user:online:%d", userID)
		pipe.Set(ctx, key, true, expire)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// DelBatchUserOnlineStatus 批量删除在线状态
func DelBatchUserOnlineStatus(ctx context.Context, userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}

	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = fmt.Sprintf("user:online:%d", userID)
	}

	return global.RDB.Del(ctx, keys...).Err()
}

// SetUserLoginStatusBatch 批量设置登录状态（用于用户登录时）
func SetUserLoginStatusBatch(ctx context.Context, tokens map[uint]string) error {
	if len(tokens) == 0 {
		return nil
	}

	pipe := global.RDB.Pipeline()
	expire := ExpireToken()

	for userID, token := range tokens {
		key := fmt.Sprintf("user:token:%d", userID)
		pipe.Set(ctx, key, token, expire)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// TokenInfo Token信息
type TokenInfo struct {
	UserID uint
	Token  string
}

// GetUserLoginStatus 获取用户登录状态
func GetUserLoginStatus(ctx context.Context, userID uint) (string, error) {
	key := fmt.Sprintf("user:token:%d", userID)
	return global.RDB.Get(ctx, key).Result()
}

// DelUserCache 删除用户缓存
func DelUserCache(ctx context.Context, userID uint) error {
	key := fmt.Sprintf("user:cache:%d", userID)
	return global.RDB.Del(ctx, key).Err()
}

// Ping 检查Redis连接
func Ping(ctx context.Context) error {
	return global.RDB.Ping(ctx).Err()
}
