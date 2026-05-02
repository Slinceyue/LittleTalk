package cache

import (
	"LittleTalk/global"
	"context"
	"fmt"
	"strconv"
)

func SetFriendList(ctx context.Context, userID uint, friendIDs []uint) error {
	key := fmt.Sprintf("friend:list:%d", userID)

	if len(friendIDs) == 0 {
		return global.RDB.Del(ctx, key).Err()
	}

	if err := global.RDB.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del failed: %w", err)
	}

	members := make([]interface{}, len(friendIDs))
	for i, id := range friendIDs {
		members[i] = id
	}

	err := global.RDB.SAdd(ctx, key, members...).Err()
	if err != nil {
		return fmt.Errorf("redis sadd failed: %w", err)
	}

	err = global.RDB.Expire(ctx, key, ExpireToken()).Err()
	if err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}

	return nil
}
func DelFriendList(ctx context.Context, userID uint) error {
	key := fmt.Sprintf("friend:list:%d", userID)
	err := global.RDB.Del(ctx, key).Err()
	return err
}
func GetFriendList(ctx context.Context, userID uint) ([]uint, error) {
	key := fmt.Sprintf("friend:list:%d", userID)

	exists, err := global.RDB.Exists(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis exists failed: %w", err)
	}
	if exists == 0 {
		return nil, fmt.Errorf("cache miss")
	}

	friendIDStrs, err := global.RDB.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis smembers failed: %w", err)
	}

	friendIDs := make([]uint, 0, len(friendIDStrs))
	for _, s := range friendIDStrs {
		id64, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			continue
		}
		friendIDs = append(friendIDs, uint(id64))
	}

	return friendIDs, nil
}
func SetFriendRequest(ctx context.Context, toUserID uint, fromUserID uint) (bool, error) {
	key := fmt.Sprintf("friend:request:%d", toUserID)

	added, err := global.RDB.SAdd(ctx, key, fromUserID).Result()
	if err != nil {
		return false, fmt.Errorf("sadd failed: %w", err)
	}

	// 设置过期（不管有没有，统一刷新，避免缓存永久不过期）
	_ = global.RDB.Expire(ctx, key, ExpireToken()).Err()

	return added > 0, nil
}

func GetFriendRequest(ctx context.Context, userID uint) ([]uint, error) {
	key := fmt.Sprintf("friend:request:%d", userID)

	// 正确：只读取，不弹出、不删除
	friendIDStrs, err := global.RDB.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis smembers failed: %w", err)
	}

	friendIDs := make([]uint, 0, len(friendIDStrs))
	for _, s := range friendIDStrs {
		id64, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			continue
		}
		friendIDs = append(friendIDs, uint(id64))
	}

	return friendIDs, nil
}
func DelFriendRequest(ctx context.Context, toUserID uint, fromUserID uint) error {
	key := fmt.Sprintf("friend:request:%d", toUserID)
	_, err := global.RDB.SRem(ctx, key, fromUserID).Result()
	return err
}
