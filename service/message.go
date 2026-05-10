package service

import (
	"LittleTalk/api/request"
	"LittleTalk/cache"
	"LittleTalk/dao"
	"LittleTalk/global"
	"LittleTalk/models"
	"LittleTalk/utils/ws"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// MessageChannel 消息发送通道
var MessageChannel = make(chan *request.MessageContext, 10000)
var senderCtx, senderCancel = context.WithCancel(context.Background())

// DbMessageChannel 异步写库通道
var DbMessageChannel = make(chan DbMessage, 5000)
var dbWriterCtx, dbWriterCancel = context.WithCancel(context.Background())

// DbMessage DB写库消息结构
type DbMessage struct {
	FromID      uint
	ToID        uint
	RoomID      uint
	MessageType int8
	Content     string
	SendTime    int64
}

// maxConcurrentSenders 限制最大并发发送数
const maxConcurrentSenders = 200

// 用于控制并发数量
var senderSemaphore = make(chan struct{}, maxConcurrentSenders)

func Run() {
	// 启动多个协程处理消息发送
	for i := 0; i < 10; i++ {
		go MessageSender()
	}
	// 启动异步写库协程
	go MessageDbWriter()
}

// MessageDbWriter 异步写库消费者
func MessageDbWriter() {
	for {
		select {
		case <-dbWriterCtx.Done():
			return
		case m, ok := <-DbMessageChannel:
			if !ok {
				return
			}
			msg := &models.Message{
				FromID:      m.FromID,
				ToID:        m.ToID,
				RoomID:      m.RoomID,
				MessageType: m.MessageType,
				Content:     m.Content,
			}
			if err := global.DB.Create(msg).Error; err != nil {
				logrus.Errorf("[DbWriter] 消息入库失败: from=%d, to=%d, room=%d, err=%v",
					m.FromID, m.ToID, m.RoomID, err)
			}
		}
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
				// 信号量已满，消息仍在 channel 中等待
				logrus.Warn("[MessageSender] 并发已满，消息等待中")
				continue
			}
			go func(m *request.MessageContext) {
				defer func() { <-senderSemaphore }()
				
				var err error
				switch m.MsgType {
				case "group_talk":
					err = SendGroupMessage(*m)
				case "talk", "friend":
					err = SendPrivateMessage(*m)
				default:
					err = errors.New("未知消息类型: " + m.MsgType)
				}
				
				if err != nil {
					logrus.Errorf("[MessageSender] 发送消息失败: type=%s, err=%v", m.MsgType, err)
				}
			}(msg)
		}
	}
}

// 停服时调用
func StopMessageSender() {
	senderCancel()
	dbWriterCancel()
	close(MessageChannel)
	close(DbMessageChannel)
}

// IsUserOnline 检查用户是否在线
func IsUserOnline(userID uint) bool {
	return ws.ConnManager.IsOnline(userID)
}

// SendPrivateMessage 发送私聊消息
func SendPrivateMessage(msg request.MessageContext) error {
	var toID, fromID uint
	var content string
	var msgType int
	var fileURL, fileName, msgID string

	switch msg.MsgType {
	case "friend":
		toID = msg.FriendMessageRequest.ToID
		fromID = msg.FriendMessageRequest.FromID
	case "talk":
		toID = msg.TalkMessageRequest.ToID
		fromID = msg.TalkMessageRequest.FromID
		content = msg.TalkMessageRequest.Content
		msgType = int(msg.TalkMessageRequest.MessageType)
		fileURL = msg.TalkMessageRequest.FileURL
		fileName = msg.TalkMessageRequest.FileName
		msgID = msg.TalkMessageRequest.MsgID
	default:
		return errors.New("未知消息类型: " + msg.MsgType)
	}

	// 验证参数
	if toID == 0 {
		return errors.New("目标用户ID不能为空")
	}
	if fromID == 0 {
		return errors.New("发送者ID不能为空")
	}

	// 好友关系检查（仅对私聊消息，不拦截好友请求通知）
	if msg.MsgType == "talk" {
		isFriend, err := dao.IsFriend(context.Background(), fromID, toID)
		if err != nil {
			logrus.Errorf("[SendPrivate] 好友关系检查失败: from=%d, to=%d, err=%v", fromID, toID, err)
			return fmt.Errorf("好友关系检查失败: %w", err)
		}
		if !isFriend {
			logrus.Warnf("[SendPrivate] 非好友发送消息被拦截: from=%d, to=%d", fromID, toID)
			return fmt.Errorf("不是好友关系，无法发送消息")
		}
	}

	logrus.Infof("[SendPrivate] from=%d, to=%d, content=%s", fromID, toID, content)

	// 异步写库（仅对实际聊天消息，不写好友请求通知）
	if msg.MsgType == "talk" {
		DbMessageChannel <- DbMessage{
			FromID:      fromID,
			ToID:        toID,
			RoomID:      0,
			MessageType: int8(msgType),
			Content:     content,
		}
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
	if msgID != "" {
		sendData["msg_id"] = msgID
	}
	if fileURL != "" {
		sendData["file_url"] = fileURL
		sendData["file_name"] = fileName
	}

	sender := map[string]interface{}{
		"msg_type": msg.MsgType,
		"data":     sendData,
	}

	// 检查目标用户是否在线
	client, ok := ws.ConnManager.Get(toID)
	if !ok || client == nil {
		// 离线消息
		if content != "" || fileURL != "" {
			if err := cache.SaveOfflineMessage(context.Background(), toID, sendData); err != nil {
				logrus.Errorf("[SendPrivate] 保存离线消息失败: %v", err)
				return err
			}
			logrus.Infof("[SendPrivate] 用户 %d 离线，消息已存入离线队列", toID)
		}
		return nil
	}

	// 在线用户，发送消息
	return writeWebSocketMessage(client, sender, toID, msgID)
}

// writeWebSocketMessage 写入WebSocket消息
func writeWebSocketMessage(client *ws.Client, data interface{}, userID uint, msgID string) error {
	client.Wmu.Lock()
	defer client.Wmu.Unlock()

	if client.Conn == nil {
		ws.ConnManager.Delete(userID)
		return errors.New("连接已断开")
	}

	// 设置5秒写入超时
	client.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	if err := client.Conn.WriteJSON(data); err != nil {
		logrus.Warnf("[WriteWS] 写入失败 UID=%d: %v", userID, err)
		_ = client.Conn.Close()
		ws.ConnManager.Delete(userID)
		
		// 尝试保存离线消息
		if msg, ok := data.(map[string]interface{}); ok {
			if dataField, hasData := msg["data"].(map[string]interface{}); hasData {
				if content, hasContent := dataField["content"].(string); hasContent && content != "" {
					_ = cache.SaveOfflineMessage(context.Background(), userID, dataField)
				}
			}
		}
		return err
	}

	logrus.Infof("[WriteWS] 发送成功 UID=%d", userID)
	return nil
}

// SendGroupMessage 发送群聊消息
func SendGroupMessage(msg request.MessageContext) error {
	roomID := msg.TalkMessageRequest.RoomID
	fromID := msg.TalkMessageRequest.FromID
	content := msg.TalkMessageRequest.Content
	msgType := msg.TalkMessageRequest.MessageType

	// 验证参数
	if roomID == 0 {
		return errors.New("群ID不能为空")
	}
	if fromID == 0 {
		return errors.New("发送者ID不能为空")
	}
	if content == "" {
		return errors.New("消息内容为空")
	}

	ctx := context.Background()

	// 先检查群是否存在
	room, err := dao.GetRoomByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("群不存在或已解散: room_id=%d", roomID)
	}

	// 检查用户是否为群成员
	if !dao.IsRoomMember(ctx, roomID, fromID) {
		// 群主自动加入
		if room.OwnerID == fromID {
			if err := dao.AddRoomMember(ctx, roomID, fromID, 1); err != nil {
				return fmt.Errorf("自动加入群失败: %v", err)
			}
			logrus.Infof("[SendGroup] 群主自动加入群: room=%d, user=%d", roomID, fromID)
		} else {
			return errors.New("不是群成员，无法发送消息")
		}
	}

	// 获取发送者信息
	sender, err := dao.GetByID(ctx, models.User{}, fromID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}

	// 异步写库
	DbMessageChannel <- DbMessage{
		FromID:      fromID,
		RoomID:      roomID,
		MessageType: int8(msgType),
		Content:     content,
	}

	sendTime := time.Now().Unix()
	avatar := sender.Avatar
	if avatar == "" {
		avatar = fmt.Sprintf("/static/avatar/%d.jpg", sender.ID)
	}

	// 构建发送数据
	sendData := map[string]interface{}{
		"from_id":       fromID,
		"from_name":     sender.Username,
		"from_avatar":   avatar,
		"room_id":       roomID,
		"content":       content,
		"message_type":  msgType,
		"send_time":     sendTime,
	}

	wsData := map[string]interface{}{
		"msg_type": "group_talk",
		"data":     sendData,
	}

	// 获取所有群成员并发送
	members, err := dao.GetRoomMembers(ctx, roomID)
	if err != nil {
		return fmt.Errorf("获取群成员失败: %v", err)
	}

	logrus.Infof("[SendGroup] room=%d, from=%d, content=%s, members=%d", roomID, fromID, content, len(members))

	sendCount := 0
	for _, m := range members {
		// 不发给自己（已在客户端直接显示）
		if m.UserID == fromID {
			continue
		}

		// 检查是否在线
		if ws.ConnManager.IsOnline(m.UserID) {
			client, _ := ws.ConnManager.Get(m.UserID)
			if client != nil {
				if err := writeWebSocketMessage(client, wsData, m.UserID, ""); err != nil {
					// 写入失败，保存离线消息
					_ = cache.SaveOfflineMessage(ctx, m.UserID, sendData)
				} else {
					sendCount++
				}
			}
		} else {
			// 离线消息存入Redis
			_ = cache.SaveOfflineMessage(ctx, m.UserID, sendData)
		}
	}

	logrus.Infof("[SendGroup] 发送完成: room=%d, 在线发送=%d, 总成员=%d", roomID, sendCount, len(members))
	return nil
}

// GetFriendListSimple 获取好友列表
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

// GetUsersInfoByIDs 批量获取用户信息
func GetUsersInfoByIDs(ctx context.Context, userIDs []uint) ([]models.User, error) {
	if len(userIDs) == 0 {
		return []models.User{}, nil
	}
	var users []models.User
	err := dao.GetByIDs(ctx, &users, userIDs)
	return users, err
}

// GetDBChatHistory 从数据库获取私聊聊天记录
func GetDBChatHistory(ctx context.Context, userID, friendID uint, page, pageSize int) ([]map[string]interface{}, error) {
	var messages []models.Message
	offset := (page - 1) * pageSize
	err := global.DB.Where(
		"(from_id = ? AND to_id = ?) OR (from_id = ? AND to_id = ?)",
		userID, friendID, friendID, userID,
	).Order("id DESC").Offset(offset).Limit(pageSize).Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// 逆序（最早的在前）
	result := make([]map[string]interface{}, 0, len(messages))
	for i := len(messages) - 1; i >= 0; i-- {
		m := messages[i]
		result = append(result, map[string]interface{}{
			"msg_id":       m.ID,
			"from_id":      m.FromID,
			"to_id":        m.ToID,
			"content":      m.Content,
			"message_type": m.MessageType,
			"send_time":    m.CreatedAt.Unix(),
		})
	}
	return result, nil
}

// SearchMessages 搜索消息（Redis缓存 + DB群消息）
func SearchMessages(ctx context.Context, userID, friendID uint, query string, page, pageSize int) ([]map[string]any, int64, error) {
	var allResults []map[string]any

	if friendID > 0 {
		// Search within specific private conversation
		msgs, err := cache.GetChatMessages(ctx, userID, friendID, 100)
		if err != nil {
			return nil, 0, err
		}
		for _, m := range msgs {
			if strings.Contains(m.Content, query) {
				allResults = append(allResults, map[string]any{
					"msg_id":      0,
					"from_id":     m.FromID,
					"to_id":       m.ToID,
					"content":     m.Content,
					"send_time":   m.SendTime,
					"message_type": m.MessageType,
				})
			}
		}
	} else {
		// Scan all private conversations
		results, err := cache.SearchUserMessages(ctx, userID, query, 200)
		if err != nil {
			return nil, 0, err
		}
		allResults = results
	}

	// Sort by send_time descending
	sort.Slice(allResults, func(i, j int) bool {
		ti, _ := allResults[i]["send_time"].(int64)
		tj, _ := allResults[j]["send_time"].(int64)
		return ti > tj
	})

	total := int64(len(allResults))

	// Paginate
	offset := (page - 1) * pageSize
	if offset >= len(allResults) {
		return []map[string]any{}, total, nil
	}
	end := offset + pageSize
	if end > len(allResults) {
		end = len(allResults)
	}

	paged := allResults[offset:end]

	// Enrich with user names
	userIDs := make(map[uint]bool)
	for _, m := range paged {
		if fid, ok := m["from_id"].(uint); ok && fid != userID {
			userIDs[fid] = true
		}
	}
	var uidList []uint
	for id := range userIDs {
		uidList = append(uidList, id)
	}
	if len(uidList) > 0 {
		var users []models.User
		if err := dao.GetByIDs(ctx, &users, uidList); err == nil {
			for _, u := range users {
				for _, m := range paged {
					if fid, ok := m["from_id"].(uint); ok && fid == u.ID {
						m["from_name"] = u.Username
						avatar := u.Avatar
						if avatar == "" {
							avatar = fmt.Sprintf("/static/avatar/%d.jpg", u.ID)
						}
						m["from_avatar"] = avatar
					}
				}
			}
		}
	}

	return paged, total, nil
}
