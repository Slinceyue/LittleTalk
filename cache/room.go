package cache

import (
	"LittleTalk/global"
	"context"
	"fmt"
	"time"
)

const (
	roomMembersKey = "room:members:%d" // 群成员缓存 KEY
)

// GetRoomMembersCache 获取群成员缓存
func GetRoomMembersCache(ctx context.Context, roomID uint) (string, error) {
	key := fmt.Sprintf(roomMembersKey, roomID)
	return global.RDB.Get(ctx, key).Result()
}

// SetRoomMembersCache 设置群成员缓存
func SetRoomMembersCache(ctx context.Context, roomID uint, data string) error {
	key := fmt.Sprintf(roomMembersKey, roomID)
	return global.RDB.Set(ctx, key, data, 5*time.Minute).Err()
}

// DelRoomMembersCache 删除群成员缓存
func DelRoomMembersCache(ctx context.Context, roomID uint) error {
	key := fmt.Sprintf(roomMembersKey, roomID)
	return global.RDB.Del(ctx, key).Err()
}
