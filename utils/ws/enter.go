package ws

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"LittleTalk/cache"
	"LittleTalk/global"

	"github.com/gorilla/websocket"
)

// ============================================================================
// 配置常量获取
// ============================================================================

// GetHeartbeatTimeout 获取心跳超时时间（客户端超过此时间无消息则判定断开）
func GetHeartbeatTimeout() time.Duration {
	if global.Config != nil && global.Config.WebSocket.HeartbeatTimeout > 0 {
		return time.Duration(global.Config.WebSocket.HeartbeatTimeout) * time.Second
	}
	return 60 * time.Second // 默认值：60秒（增加容错性）
}

func getReadBufferSize() int {
	if global.Config != nil && global.Config.WebSocket.ReadBufferSize > 0 {
		return global.Config.WebSocket.ReadBufferSize
	}
	return 1024 // 默认值
}

func getWriteBufferSize() int {
	if global.Config != nil && global.Config.WebSocket.WriteBufferSize > 0 {
		return global.Config.WebSocket.WriteBufferSize
	}
	return 1024 // 默认值
}

// ============================================================================
// 核心数据结构
// ============================================================================

// ConnectionManager WS连接管理器
// 使用读写锁实现并发安全：读操作共享锁，写操作互斥锁
type ConnectionManager struct {
	sync.RWMutex              // 读写锁：读共享，写独占
	conns map[uint]*Client    // 连接映射：userID -> Client
}

// Client WebSocket客户端连接
type Client struct {
	UserID uint              // 用户ID
	Conn   *websocket.Conn   // WebSocket连接
	Wmu    sync.Mutex        // 连接级写锁（保护并发写入）
}

// ============================================================================
// 全局连接管理器初始化
// ============================================================================

var ConnManager = &ConnectionManager{
	conns: make(map[uint]*Client),
}

// ============================================================================
// 连接管理核心方法
// ============================================================================

// Add 添加连接（WS建立时调用）
// 线程安全：使用写锁保护连接映射
func (m *ConnectionManager) Add(userID uint, conn *websocket.Conn) {
	m.Lock()
	defer m.Unlock()

	// 检查是否存在旧连接，防止同一用户多设备登录
	if oldClient, exists := m.conns[userID]; exists {
		log.Printf("[WS] [UID=%d] 检测到旧连接，已关闭旧连接\n", userID)
		oldClient.Conn.Close()
	}

	// 添加新连接到管理器
	m.conns[userID] = &Client{
		UserID: userID,
		Conn:   conn,
	}

	// 设置Redis在线状态（永久有效，由WS连接生命周期管理）
	ctx := context.Background()
	if err := cache.SetUserOnlineStatusPermanent(ctx, userID); err != nil {
		log.Printf("[WS] [UID=%d] 设置Redis在线状态失败: %v\n", userID, err)
	}

	log.Printf("[WS] [UID=%d] 连接已建立，当前在线人数: %d\n", userID, len(m.conns))
}

// Get 获取连接（使用读锁，并发安全）
// 注意：返回的Client对象被Caller持有，外部goroutine使用时需注意生命周期
func (m *ConnectionManager) Get(userID uint) (conn *Client, ok bool) {
	m.RLock()
	defer m.RUnlock()
	conn, ok = m.conns[userID]
	return
}

// IsOnline 检查用户是否在线（直接查询内存，O(1)复杂度）
func (m *ConnectionManager) IsOnline(userID uint) bool {
	m.RLock()
	defer m.RUnlock()
	_, ok := m.conns[userID]
	return ok
}

// Delete 删除连接（WS断开时调用）
// 线程安全：使用写锁保护连接映射
func (m *ConnectionManager) Delete(userID uint) {
	m.Lock()

	// 检查连接是否存在（防止重复删除）
	if _, exists := m.conns[userID]; !exists {
		m.Unlock()
		return
	}

	// 从内存中移除连接
	delete(m.conns, userID)

	// 删除Redis在线状态
	ctx := context.Background()
	if err := cache.DelUserOnlineStatus(ctx, userID); err != nil {
		log.Printf("[WS] [UID=%d] 删除Redis在线状态失败: %v\n", userID, err)
	}

	currentCount := len(m.conns)
	m.Unlock()

	log.Printf("[WS] [UID=%d] 连接已断开，当前在线人数: %d\n", userID, currentCount)

	// 异步广播离线状态（避免阻塞断开流程）
	go m.BroadcastOnlineStatus(userID, false)
}

// GetAllOnlineUsers 获取所有在线用户ID列表
func (m *ConnectionManager) GetAllOnlineUsers() []uint {
	m.RLock()
	defer m.RUnlock()
	users := make([]uint, 0, len(m.conns))
	for id := range m.conns {
		users = append(users, id)
	}
	return users
}

// Count 获取当前在线人数
func (m *ConnectionManager) Count() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.conns)
}

// ============================================================================
// 广播功能
// ============================================================================

// BroadcastOnlineStatus 广播用户在线状态变化给所有在线用户
// 注意：此方法由持有读锁的goroutine调用，解锁后才启动异步广播
func (m *ConnectionManager) BroadcastOnlineStatus(userID uint, online bool) {
	// 先获取快照，避免长时间持有读锁
	m.RLock()
	onlineStatus := map[string]any{
		"type":    "online_status",
		"user_id": userID,
		"online":  online,
	}

	// 复制连接列表用于遍历
	clients := make(map[uint]*Client, len(m.conns))
	for uid, client := range m.conns {
		clients[uid] = client
	}
	m.RUnlock()

	// 向除自己外的所有在线用户广播
	for uid, client := range clients {
		if uid == userID {
			continue
		}

		client.Wmu.Lock()
		client.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		err := client.Conn.WriteJSON(onlineStatus)
		client.Wmu.Unlock()

		if err != nil {
			log.Printf("[WS] [UID=%d] 广播在线状态[UID=%d,online=%v]失败: %v\n", uid, userID, online, err)
		}
	}
}

// ============================================================================
// WebSocket Upgrader
// ============================================================================

// GetUpgrader 获取Upgrader实例（用于HTTP升级为WebSocket）
func GetUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  getReadBufferSize(),
		WriteBufferSize: getWriteBufferSize(),
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}
