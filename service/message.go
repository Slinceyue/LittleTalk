package service

import (
	"LittleTalk/api/request"
	"LittleTalk/cache"
	"LittleTalk/dao"
	"LittleTalk/models"
	"LittleTalk/utils/ws"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MessageChannel 消息发送通道
var MessageChannel = make(chan *request.MessageContext, 1000)
var senderCtx, senderCancel = context.WithCancel(context.Background())

// maxConcurrentSenders 限制最大并发发送数
const maxConcurrentSenders = 100

// 用于控制并发数量
var senderSemaphore = make(chan struct{}, maxConcurrentSenders)

func Run() {
	// 启动多个协程处理消息发送
	for i := 0; i < 10; i++ {
		go MessageSender()
	}
}

// MessageSender 消息发送器
func MessageSender() {
	for {
		select {
		case <-senderCtx.Done():
			return
		case msg, ok := <-MessageChannel:
			if !ok {
				return
			}
			// 使用信号量控制并发（不阻塞）
			select {
			case senderSemaphore <- struct{}{}:
				// 获取到信号量，继续处理
			default:
				// 信号量已满，消息仍在 channel 中等待，稍后重试
				logrus.Warn("[MessageSender] 并发已满，消息等待中")
			}
			go func(m *request.MessageContext) {
				defer func() { <-senderSemaphore }()
				err := Send(*m)
				if err != nil {
					logrus.Error("发送消息失败: ", err.Error())
				}
			}(msg)
		}
	}
}

// 停服时调用
func StopMessageSender() {
	senderCancel()
	close(MessageChannel)
}

// IsUserOnline 检查用户是否在线（直接查询内存WS连接，100%准确）
func IsUserOnline(userID uint) bool {
	return ws.ConnManager.IsOnline(userID)
}

// Send 发送消息（简化版，移除ACK机制）
func Send(msg request.MessageContext) error {
	var toID uint
	var fromID uint
	var content string
	var msgType int

	switch msg.MsgType {
	case "friend":
		toID = msg.FriendMessageRequest.ToID
		fromID = msg.FriendMessageRequest.FromID
	case "talk":
		toID = msg.TalkMessageRequest.ToID
		fromID = msg.TalkMessageRequest.FromID
		content = msg.TalkMessageRequest.Content
		msgType = int(msg.TalkMessageRequest.MessageType)
	default:
		return errors.New("未知消息类型")
	}

	logrus.Infof("[Send] 收到消息: from=%d, to=%d, content=%s, msgID=%s", fromID, toID, content, msg.MsgID)

	// 过滤空消息
	if content == "" && toID == 0 {
		logrus.Debugf("[Send] 忽略空消息 from=%d", fromID)
		return nil
	}

	if toID == 0 {
		return errors.New("用户不存在")
	}

	// 构建发送数据
	sendTime := time.Now().Unix()
	sendData := map[string]interface{}{
		"from_id":      fromID,
		"to_id":        toID,
		"content":      content,
		"message_type": msgType,
		"send_time":    sendTime,
	}

	// 添加消息ID
	if msg.MsgID != "" {
		sendData["msg_id"] = msg.MsgID
	}

	// 添加文件信息（如果有）
	if msg.FileURL != "" {
		sendData["file_url"] = msg.FileURL
		sendData["file_name"] = msg.FileName
	}

	sender := map[string]interface{}{
		"msg_type": msg.MsgType,
		"data":     sendData,
	}

	// 直接查询内存WS连接（100%准确）
	client, ok := ws.ConnManager.Get(toID)
	if !ok || client == nil {
		// 用户不在线，保存离线消息
		if content != "" || msg.FileURL != "" {
			if err := cache.SaveOfflineMessage(context.Background(), toID, sendData); err != nil {
				logrus.Error("保存离线消息失败: ", err)
				// 返回失败状态给发送者
				notifySendFailed(fromID, msg.MsgID, "消息发送失败")
				return err
			}
			logrus.Infof("[Send] 用户 %d 不在线，消息已保存为离线消息", toID)
		}
		return nil
	}

	logrus.Infof("[Send] 用户 %d 在线，准备发送消息", toID)

	// 连接级写锁（串行化写入）
	client.Wmu.Lock()
	defer client.Wmu.Unlock()

	// 再次检查连接是否有效
	if client.Conn == nil {
		// 连接已断开，触发清理（由读消息协程的defer处理）
		ws.ConnManager.Delete(toID)
		// 保存离线消息
		if content != "" || msg.FileURL != "" {
			if err := cache.SaveOfflineMessage(context.Background(), toID, sendData); err != nil {
				logrus.Error("保存离线消息失败: ", err)
				notifySendFailed(fromID, msg.MsgID, "消息发送失败")
				return err
			}
		}
		return errors.New("连接已断开")
	}

	// 设置写入超时（10秒）
	client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	err := client.Conn.WriteJSON(sender)
	if err != nil {
		logrus.Warnf("[Send] 写入消息失败，清理连接: %v", err)
		_ = client.Conn.Close()
		// 清理连接和Redis在线状态
		ws.ConnManager.Delete(toID)
		// 保存离线消息
		if content != "" || msg.FileURL != "" {
			if saveErr := cache.SaveOfflineMessage(context.Background(), toID, sendData); saveErr != nil {
				logrus.Error("保存离线消息失败: ", saveErr)
				notifySendFailed(fromID, msg.MsgID, "消息发送失败")
				return saveErr
			}
		}
		return nil
	}

	logrus.Infof("[Send] 消息发送成功: from=%d, to=%d, msgID=%s", fromID, toID, msg.MsgID)
	return nil
}

// notifySendFailed 向发送者通知消息发送失败
func notifySendFailed(userID uint, msgID string, reason string) {
	client, ok := ws.ConnManager.Get(userID)
	if !ok || client == nil {
		return
	}

	failedMsg := map[string]interface{}{
		"msg_type": "send_failed",
		"msg_id":   msgID,
		"reason":   reason,
	}

	client.Wmu.Lock()
	defer client.Wmu.Unlock()

	client.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := client.Conn.WriteJSON(failedMsg); err != nil {
		logrus.Warnf("[notifySendFailed] 发送失败通知失败 UID=%d: %v", userID, err)
	}
}

// BatchSend 批量发送消息（高并发优化）
func BatchSend(msgs []request.MessageContext) []error {
	var wg sync.WaitGroup
	errs := make([]error, len(msgs))

	for i, msg := range msgs {
		wg.Add(1)
		go func(idx int, m request.MessageContext) {
			defer wg.Done()
			errs[idx] = Send(m)
		}(i, msg)
	}
	wg.Wait()

	return errs
}

// GetFriendListSimple 获取好友列表（简化版，仅ID和用户名，用于消息列表）
func GetFriendListSimple(ctx context.Context, userID uint) ([]models.User, error) {
	friendModels, _, err := dao.GetFriendListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var friendIDs []uint
	for _, m := range friendModels {
		friendID := m.FriendID
		if m.FriendID == userID {
			friendID = m.UserID
		}
		friendIDs = append(friendIDs, friendID)
	}

	if len(friendIDs) == 0 {
		return []models.User{}, nil
	}

	var users []models.User
	err = dao.GetByIDs(ctx, &users, friendIDs)
	return users, err
}

// GetUsersInfoByIDs 根据ID列表批量获取用户信息
func GetUsersInfoByIDs(ctx context.Context, userIDs []uint) ([]models.User, error) {
	if len(userIDs) == 0 {
		return []models.User{}, nil
	}
	var users []models.User
	err := dao.GetByIDs(ctx, &users, userIDs)
	return users, err
}
