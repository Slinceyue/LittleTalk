package message_handler

import (
	"LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/cache"
	"LittleTalk/models/enum"
	"LittleTalk/service"
	"LittleTalk/utils/ws"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// 结构体定义
// ============================================================================

// MessageHandler 消息处理器（HTTP接口 + WebSocket）
type MessageHandler struct{}

// MessageListItem 消息列表项（用于GetMessageList接口）
type MessageListItem struct {
	FriendID     uint   `json:"friend_id"`      // 好友ID
	FriendName   string `json:"friend_name"`    // 好友名称
	FriendAvatar string `json:"friend_avatar"`  // 好友头像
	LastMessage  string `json:"last_message"`  // 最后一条消息
	SendTime     int64  `json:"send_time"`      // 发送时间
	Online       bool   `json:"online"`         // 是否在线
}

// ============================================================================
// WebSocket 核心处理
// ============================================================================

// WS WebSocket连接处理函数
// 流程：建立连接 -> 注册到管理器 -> 启动协程处理离线消息和广播 -> 阻塞读取消息循环
func (MessageHandler) WS(c *gin.Context) {
	// ----- 1. 身份验证 -----
	userIDVal, exists := c.Get("id")
	if !exists {
		log.Println("[WS] 错误：未获取到用户ID")
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}
	userID := userIDVal.(uint)

	log.Printf("[WS] [UID=%d] 正在建立WebSocket连接...\n", userID)

	// ----- 2. 升级为WebSocket连接 -----
	conn, err := ws.GetUpgrader().Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] [UID=%d] 连接升级失败: %v\n", userID, err)
		return
	}

	// ----- 3. 注册到连接管理器（内部自动设置Redis在线状态） -----
	ws.ConnManager.Add(userID, conn)

	// ----- 4. 获取客户端对象用于后续操作 -----
	client, ok := ws.ConnManager.Get(userID)
	if !ok {
		log.Printf("[WS] [UID=%d] 获取客户端对象失败\n", userID)
		conn.Close()
		return
	}

	log.Printf("[WS] [UID=%d] WebSocket连接成功，当前在线用户数: %d\n", userID, ws.ConnManager.Count())

	// ----- 5. 启动异步任务 -----
	// 广播上线状态给其他在线用户
	go ws.ConnManager.BroadcastOnlineStatus(userID, true)
	// 发送离线消息
	go sendOfflineMessages(userID, client)

	// ----- 6. 设置连接断开时的清理逻辑（defer） -----
	defer func() {
		log.Printf("[WS] [UID=%d] 连接断开，触发清理流程\n", userID)
		// Delete 会：1.从内存移除连接 2.删除Redis在线状态 3.异步广播离线状态
		ws.ConnManager.Delete(userID)
		conn.Close()
	}()

	// ----- 7. 阻塞读取消息循环 -----
	// 任何错误（包含超时）都会导致循环退出，触发defer清理
	for {
		// 设置读取超时：超过心跳超时时间无消息则判定断开
		conn.SetReadDeadline(time.Now().Add(ws.GetHeartbeatTimeout()))

		// 读取消息（阻塞，直到收到消息或超时或断开）
		_, data, err := conn.ReadMessage()
		if err != nil {
			// 根据错误类型判断断开原因
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("[WS] [UID=%d] 读取消息超时（超过%v无消息）\n", userID, ws.GetHeartbeatTimeout())
			} else {
				log.Printf("[WS] [UID=%d] 读取消息失败: %v\n", userID, err)
			}
			// 连接已断开，退出循环，defer自动清理
			return
		}

		// ----- 8. 解析并处理消息 -----
		processWSMessage(userID, client, data)
	}
}

// processWSMessage 处理WebSocket消息（内部方法）
// 保证并发安全：使用Wmu锁保护写入操作
func processWSMessage(userID uint, client *ws.Client, data []byte) {
	// 解析消息类型
	var baseMsg map[string]any
	if err := json.Unmarshal(data, &baseMsg); err != nil {
		log.Printf("[WS] [UID=%d] JSON解析失败: %v\n", userID, err)
		return
	}

	msgType, _ := baseMsg["type"].(string)

	// ----- 心跳：客户端发ping，服务端回pong -----
	if msgType == "ping" {
		client.Wmu.Lock()
		// 设置2秒写入超时
		client.Conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		err := client.Conn.WriteJSON(map[string]any{"type": "pong"})
		client.Wmu.Unlock()
		if err != nil {
			log.Printf("[WS] [UID=%d] 回复心跳失败: %v\n", userID, err)
			// 连续心跳失败超过3次，断开连接
			client.HeartbeatTimeoutCount++
			if client.HeartbeatTimeoutCount >= 3 {
				log.Printf("[WS] [UID=%d] 心跳超时次数过多，断开连接\n", userID)
				return // 断开连接
			}
		} else {
			client.HeartbeatTimeoutCount = 0
		}
		return
	}

	// ----- 获取好友在线状态 -----
	if msgType == "get_online_status" {
		handleGetOnlineStatus(userID, client)
		return
	}

	// ----- 打字状态中继 -----
	if msgType == "typing" {
		var typingMsg struct {
			ToID   uint `json:"to_id"`
			Typing bool `json:"typing"`
		}
		if err := json.Unmarshal(data, &typingMsg); err != nil {
			return
		}
		if typingMsg.ToID > 0 {
			relayTypingIndicator(userID, typingMsg.ToID, typingMsg.Typing)
		}
		return
	}

	// ----- 已读回执中继 -----
	if msgType == "read_receipt" {
		var receiptMsg struct {
			MsgIDs []string `json:"msg_ids"`
			ToID   uint     `json:"to_id"`
		}
		if err := json.Unmarshal(data, &receiptMsg); err != nil {
			return
		}
		if receiptMsg.ToID > 0 && len(receiptMsg.MsgIDs) > 0 {
			relayReadReceipt(userID, receiptMsg.ToID, receiptMsg.MsgIDs)
		}
		return
	}

	// ----- 处理聊天消息 -----
	var message request.TalkMessageRequest
	if err := json.Unmarshal(data, &message); err != nil {
		log.Printf("[WS] [UID=%d] 解析聊天消息失败: %v\n", userID, err)
		return
	}
	message.FromID = userID

	// 如果有 room_id 且为群聊类型，发送到群聊消息队列
	if message.RoomID > 0 {
		service.MessageChannel <- &request.MessageContext{
			MsgType:           "group_talk",
			TalkMessageRequest: message,
		}
		return
	}

	// 如果有 to_id 且大于 0，发送到私聊消息队列
	if message.ToID > 0 {
		service.MessageChannel <- &request.MessageContext{
			MsgType:           "talk",
			TalkMessageRequest: message,
		}
		return
	}

	log.Printf("[WS] [UID=%d] 消息缺少目标ID (to_id 或 room_id)\n", userID)
}

// handleGetOnlineStatus 处理获取好友在线状态请求
func handleGetOnlineStatus(userID uint, client *ws.Client) {
	friends, err := service.GetFriendListSimple(context.Background(), userID)
	if err != nil {
		log.Printf("[WS] [UID=%d] 获取好友列表失败: %v\n", userID, err)
		return
	}

	var statuses []map[string]any
	for _, f := range friends {
		statuses = append(statuses, map[string]any{
			"user_id": f.ID,
			"online":  ws.ConnManager.IsOnline(f.ID),
		})
	}

	client.Wmu.Lock()
	err = client.Conn.WriteJSON(map[string]any{
		"type":     "batch_online_status",
		"statuses": statuses,
	})
	client.Wmu.Unlock()

	if err != nil {
		log.Printf("[WS] [UID=%d] 发送在线状态失败: %v\n", userID, err)
	}
}

// sendOfflineMessages 发送离线消息给用户（异步执行）
// 发送失败时自动终止，避免无效操作
func sendOfflineMessages(userID uint, client *ws.Client) {
	ctx := context.Background()
	messages, err := cache.GetOfflineMessages(ctx, userID)
	if err != nil {
		log.Printf("[离线消息] [UID=%d] 获取失败: %v\n", userID, err)
		return
	}

	if len(messages) == 0 {
		return
	}

	log.Printf("[离线消息] [UID=%d] 有 %d 条离线消息待发送\n", userID, len(messages))

	for i, rawMsg := range messages {
		var msg map[string]any
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			log.Printf("[离线消息] [UID=%d] 第%d条JSON解析失败，跳过\n", userID, i+1)
			continue
		}

		// 根据消息内容判断类型：room_id > 0 表示群聊消息
		msgType := "talk"
		if roomID, ok := msg["room_id"].(float64); ok && roomID > 0 {
			msgType = "group_talk"
		}

		client.Wmu.Lock()
		err = client.Conn.WriteJSON(map[string]any{
			"msg_type": msgType,
			"data":     msg,
		})
		client.Wmu.Unlock()

		if err != nil {
			log.Printf("[离线消息] [UID=%d] 第%d条发送失败，终止离线消息推送: %v\n", userID, i+1, err)
			return
		}
	}

	log.Printf("[离线消息] [UID=%d] 离线消息已全部发送\n", userID)
}

// ============================================================================
// HTTP 接口
// ============================================================================

// GetMessageList 获取消息列表（最近聊天）
// 返回每个好友的最后一条消息及基本信息
func (MessageHandler) GetMessageList(c *gin.Context) {
	userID, _ := c.Get("id")
	ctx := c.Request.Context()

	// 获取最近聊天记录
	messages, err := cache.GetRecentChats(ctx, userID.(uint))
	if err != nil {
		log.Printf("[消息列表] [UID=%d] 获取聊天记录失败: %v\n", userID, err)
		response.FailWithMsg(c, enum.CodeServerError, "获取消息列表失败")
		return
	}

	// 获取好友列表（用于显示好友名称和头像）
	friends, err := service.GetFriendListSimple(ctx, userID.(uint))
	if err != nil {
		log.Printf("[消息列表] [UID=%d] 获取好友列表失败: %v\n", userID, err)
	}

	// 构建好友ID到名称/头像的映射
	friendMap := make(map[uint]string)
	friendAvatarMap := make(map[uint]string)
	for _, f := range friends {
		friendMap[f.ID] = f.Username
		avatar := f.Avatar
		if avatar == "" {
			avatar = fmt.Sprintf("/static/avatar/%d.jpg", f.ID)
		}
		friendAvatarMap[f.ID] = avatar
	}

	// 构建消息列表
	var result []MessageListItem
	seen := make(map[uint]bool)

	for _, msg := range messages {
		var friendID uint
		if msg.FromID == userID.(uint) {
			friendID = msg.ToID
		} else {
			friendID = msg.FromID
		}

		// 避免重复
		if seen[friendID] {
			continue
		}
		seen[friendID] = true

		friendName := friendMap[friendID]
		if friendName == "" {
			friendName = "用户" + strconv.FormatUint(uint64(friendID), 10)
		}
		friendAvatar := friendAvatarMap[friendID]
		if friendAvatar == "" {
			friendAvatar = fmt.Sprintf("/static/avatar/%d.jpg", friendID)
		}

		result = append(result, MessageListItem{
			FriendID:     friendID,
			FriendName:   friendName,
			FriendAvatar: friendAvatar,
			LastMessage:  msg.Content,
			SendTime:     msg.SendTime,
			Online:       ws.ConnManager.IsOnline(friendID),
		})
	}

	response.OKWithData(c, result)
}

// GetDBChatHistory 从数据库获取聊天记录
func (MessageHandler) GetDBChatHistory(c *gin.Context) {
	userID, _ := c.Get("id")
	friendIDStr := c.Query("friend_id")
	if friendIDStr == "" {
		response.FailWithMsg(c, enum.CodeInvalidParam, "缺少friend_id参数")
		return
	}

	fid, err := strconv.ParseUint(friendIDStr, 10, 64)
	if err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "无效的friend_id")
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "50")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	ctx := c.Request.Context()
	messages, err := service.GetDBChatHistory(ctx, userID.(uint), uint(fid), page, pageSize)
	if err != nil {
		log.Printf("[DB聊天记录] [UID=%d] 获取与[UID=%d]的记录失败: %v\n", userID, fid, err)
		response.FailWithMsg(c, enum.CodeServerError, "获取聊天记录失败")
		return
	}

	response.OKWithData(c, messages)
}

// GetChatHistory 获取与指定好友的聊天记录（Redis缓存）
func (MessageHandler) GetChatHistory(c *gin.Context) {
	userID, _ := c.Get("id")
	friendIDStr := c.Query("friend_id")
	if friendIDStr == "" {
		response.FailWithMsg(c, enum.CodeInvalidParam, "缺少friend_id参数")
		return
	}

	fid, err := strconv.ParseUint(friendIDStr, 10, 64)
	if err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "无效的friend_id")
		return
	}

	ctx := c.Request.Context()
	messages, err := cache.GetChatMessages(ctx, userID.(uint), uint(fid), 50)
	if err != nil {
		log.Printf("[聊天记录] [UID=%d] 获取与[UID=%d]的聊天记录失败: %v\n", userID, fid, err)
		response.FailWithMsg(c, enum.CodeServerError, "获取聊天记录失败")
		return
	}

	response.OKWithData(c, messages)
}

// GetUnreadCount 获取未读消息数量
func (MessageHandler) GetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("id")
	ctx := c.Request.Context()

	count, err := cache.GetUnreadCount(ctx, userID.(uint))
	if err != nil {
		log.Printf("[未读消息] [UID=%d] 获取未读消息数失败: %v\n", userID, err)
		response.FailWithMsg(c, enum.CodeServerError, "获取未读消息数失败")
		return
	}

	response.OKWithData(c, gin.H{"total": count})
}

// SearchMessages 搜索消息
func (MessageHandler) SearchMessages(c *gin.Context) {
	userID, _ := c.Get("id")
	query := c.Query("query")
	if query == "" {
		response.FailWithMsg(c, enum.CodeInvalidParam, "缺少搜索关键词")
		return
	}

	friendIDStr := c.Query("friend_id")
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")

	fid := uint(0)
	if friendIDStr != "" {
		parsed, err := strconv.ParseUint(friendIDStr, 10, 64)
		if err == nil {
			fid = uint(parsed)
		}
	}

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 50 {
			pageSize = ps
		}
	}

	ctx := c.Request.Context()
	list, total, err := service.SearchMessages(ctx, userID.(uint), fid, query, page, pageSize)
	if err != nil {
		log.Printf("[消息搜索] [UID=%d] 搜索失败: %v\n", userID, err)
		response.FailWithMsg(c, enum.CodeServerError, "搜索失败")
		return
	}

	response.OKWithData(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// relayTypingIndicator 转发打字状态给目标用户
func relayTypingIndicator(fromID, toID uint, typing bool) {
	target, ok := ws.ConnManager.Get(toID)
	if !ok || target == nil {
		return
	}

	target.Wmu.Lock()
	defer target.Wmu.Unlock()

	if target.Conn == nil {
		return
	}

	target.Conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if err := target.Conn.WriteJSON(map[string]any{
		"type":    "typing",
		"from_id": fromID,
		"typing":  typing,
	}); err != nil {
		log.Printf("[Typing] 转发打字状态失败 from=%d to=%d: %v\n", fromID, toID, err)
	}
}

// relayReadReceipt 转发已读回执给消息发送者
func relayReadReceipt(fromID, toID uint, msgIDs []string) {
	target, ok := ws.ConnManager.Get(toID)
	if !ok || target == nil {
		return
	}

	target.Wmu.Lock()
	defer target.Wmu.Unlock()

	if target.Conn == nil {
		return
	}

	target.Conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if err := target.Conn.WriteJSON(map[string]any{
		"type":    "read_receipt",
		"from_id": fromID,
		"msg_ids": msgIDs,
	}); err != nil {
		log.Printf("[ReadReceipt] 转发已读回执失败 from=%d to=%d: %v\n", fromID, toID, err)
	}
}
