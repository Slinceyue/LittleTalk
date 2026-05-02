package ws

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type ConnectionManager struct {
	sync.RWMutex // 读写锁：读共享，写独占
	conns        map[uint]*Client
}
type Client struct {
	Conn *websocket.Conn
	Wmu  sync.Mutex
}

// 初始化
var ConnManager = &ConnectionManager{
	conns: make(map[uint]*Client),
}

func (m *ConnectionManager) Add(userID uint, conn *websocket.Conn) {
	m.Lock()
	defer m.Unlock()
	m.conns[userID] = &Client{Conn: conn}
}
func (m *ConnectionManager) Get(userID uint) (conn *Client, ok bool) {
	m.Lock()
	defer m.Unlock()
	conn, ok = m.conns[userID]
	return
}
func (m *ConnectionManager) Delete(userID uint) {
	m.Lock()
	defer m.Unlock()
	delete(m.conns, userID)
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 跨域放行，前端本地调试必开
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
