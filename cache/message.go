package cache

import (
	"LittleTalk/global"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
)

const (
	unreadMsgKey    = "msg:unread:%d"   // 未读消息 KEY
	chatMsgKey      = "msg:chat:%d:%d"  // 聊天消息 KEY (用户A:用户B)
	msgProcessedKey = "msg:processed"   // 已处理消息ID集合（去重）
)

// 从配置读取消息参数
func getMsgExpire() time.Duration {
	if global.Config != nil && global.Config.Message.MsgExpire > 0 {
		return time.Duration(global.Config.Message.MsgExpire) * 24 * time.Hour
	}
	return 7 * 24 * time.Hour // 默认7天
}

func getMaxChatHistory() int {
	if global.Config != nil && global.Config.Message.MaxChatHistory > 0 {
		return global.Config.Message.MaxChatHistory
	}
	return 100 // 默认100条
}

func getMsgProcessedTTL() time.Duration {
	if global.Config != nil && global.Config.Message.MsgProcessedTTL > 0 {
		return time.Duration(global.Config.Message.MsgProcessedTTL) * time.Minute
	}
	return 5 * time.Minute // 默认5分钟
}

// ChatMessage 聊天消息结构
type ChatMessage struct {
	FromID      uint   `json:"from_id"`
	ToID        uint   `json:"to_id"`
	Content     string `json:"content"`
	MessageType int    `json:"message_type"`
	SendTime    int64  `json:"send_time"`
}

// SaveChatMessage 保存聊天消息到Redis
func SaveChatMessage(ctx context.Context, fromID, toID uint, content string, msgType int) error {
	msg := ChatMessage{
		FromID:      fromID,
		ToID:        toID,
		Content:     content,
		MessageType: msgType,
		SendTime:    time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// 双向存储：用户A和用户B的对话记录都存储
	keys := []string{
		fmt.Sprintf(chatMsgKey, fromID, toID),
		fmt.Sprintf(chatMsgKey, toID, fromID),
	}

	for _, key := range keys {
		// 左边推入（新消息在前）
		_, err = global.RDB.LPush(ctx, key, data).Result()
		if err != nil {
			return err
		}
		// 限制消息数量
		global.RDB.LTrim(ctx, key, 0, int64(getMaxChatHistory()-1))
		// 设置过期
		global.RDB.Expire(ctx, key, getMsgExpire())
	}

	return nil
}

// GetChatMessages 获取与指定用户的聊天记录
func GetChatMessages(ctx context.Context, userID, friendID uint, limit int) ([]ChatMessage, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	key := fmt.Sprintf(chatMsgKey, userID, friendID)
	list, err := global.RDB.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	// LRange(0, n) 获取的是最新消息在前，需要反转成历史顺序（最旧在前）
	var messages []ChatMessage
	// 从后往前遍历，实现反转
	for i := len(list) - 1; i >= 0; i-- {
		var msg ChatMessage
		if err := json.Unmarshal([]byte(list[i]), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetRecentChats 获取最近聊天的好友列表（用于消息列表）
func GetRecentChats(ctx context.Context, userID uint) ([]ChatMessage, error) {
	// 扫描该用户的所有聊天记录key
	pattern := fmt.Sprintf("msg:chat:%d:*", userID)
	keys, _, err := global.RDB.Scan(ctx, 0, pattern, 100).Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return []ChatMessage{}, nil
	}

	// 使用Pipeline获取每个对话的最后一条消息
	pipe := global.RDB.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.LIndex(ctx, key, 0)
	}
	pipe.Exec(ctx)

	var messages []ChatMessage
	for _, cmd := range cmds {
		s, err := cmd.Result()
		if err != nil || s == "" {
			continue
		}
		var msg ChatMessage
		if err := json.Unmarshal([]byte(s), &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

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
	_, _ = global.RDB.Expire(ctx, key, getMsgExpire()).Result()
	return nil
}

// GetUnreadCount 获取未读消息数量
func GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	key := fmt.Sprintf(unreadMsgKey, userID)
	return global.RDB.LLen(ctx, key).Result()
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

// IsMessageProcessed 检查消息是否已处理（幂等性）
func IsMessageProcessed(msgID string) (bool, error) {
	ctx := context.Background()
	exists, err := global.RDB.SIsMember(ctx, msgProcessedKey, msgID).Result()
	if err != nil {
		return false, err
	}
	return exists, nil
}

// SearchUserMessages 搜索用户的所有私聊消息
func SearchUserMessages(ctx context.Context, userID uint, query string, maxResults int) ([]map[string]any, error) {
	pattern := fmt.Sprintf("msg:chat:%d:*", userID)
	keys, _, err := global.RDB.Scan(ctx, 0, pattern, 100).Result()
	if err != nil {
		return nil, err
	}

	if maxResults <= 0 {
		maxResults = 200
	}

	var results []map[string]any
	queryLower := strings.ToLower(query)

	for _, key := range keys {
		if len(results) >= maxResults {
			break
		}
		list, err := global.RDB.LRange(ctx, key, 0, -1).Result()
		if err != nil {
			continue
		}
		for _, s := range list {
			if len(results) >= maxResults {
				break
			}
			var msg ChatMessage
			if err := json.Unmarshal([]byte(s), &msg); err != nil {
				continue
			}
			if strings.Contains(strings.ToLower(msg.Content), queryLower) {
				results = append(results, map[string]any{
					"msg_id":       0,
					"from_id":      msg.FromID,
					"to_id":        msg.ToID,
					"content":      msg.Content,
					"send_time":    msg.SendTime,
					"message_type": msg.MessageType,
				})
			}
		}
	}

	return results, nil
}

// MarkMessageProcessed 标记消息已处理
func MarkMessageProcessed(msgID string) error {
	ctx := context.Background()
	pipe := global.RDB.Pipeline()
	pipe.SAdd(ctx, msgProcessedKey, msgID)
	pipe.Expire(ctx, msgProcessedKey, getMsgProcessedTTL())
	_, err := pipe.Exec(ctx)
	return err
}
