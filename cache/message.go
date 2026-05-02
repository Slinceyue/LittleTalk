package cache

import (
	"LittleTalk/global"
	"context"
	"fmt"
	"time"

	"github.com/goccy/go-json"
)

const (
	unreadMsgKey = "msg:unread:%d"    // 未读消息 KEY
	msgExpire    = 7 * 24 * time.Hour // 消息保存7天
)

func SaveOfflineMessage(ctx context.Context, userID uint, msg any) error {
	key := fmt.Sprintf(unreadMsgKey, userID)

	// 序列化为 JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// 1. 右边推入列表
	_, err = global.RDB.RPush(ctx, key, data).Result()
	if err != nil {
		return err
	}

	// 2. 设置过期（避免长期不用占内存）
	_, _ = global.RDB.Expire(ctx, key, msgExpire).Result()
	return nil
}

// GetOfflineMessages 获取未读消息并清空
func GetOfflineMessages(ctx context.Context, userID uint) ([]json.RawMessage, error) {
	key := fmt.Sprintf(unreadMsgKey, userID)

	// 1. 获取所有消息 0 -1 代表全部
	list, err := global.RDB.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	// 2. 清空未读（已拉取）
	_, _ = global.RDB.Del(ctx, key).Result()

	// 3. 转成数组返回
	var messages []json.RawMessage
	for _, s := range list {
		messages = append(messages, json.RawMessage(s))
	}

	return messages, nil
}
