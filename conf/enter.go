package conf

type Config struct {
	System    System    `yaml:"system"`
	Log       Log       `yaml:"log"`
	DB        DB        `yaml:"db"`   // 从库配置（读操作）
	DBMaster  DB        `yaml:"db_master"` // 主库配置（写操作）
	Jwt       Jwt       `yaml:"jwt"`
	Redis     Redis     `yaml:"redis"`
	WebSocket WebSocket `yaml:"websocket"`
	Message   Message   `yaml:"message"`
	Cache     Cache     `yaml:"cache"`
}

// WebSocket WebSocket配置
type WebSocket struct {
	HeartbeatInterval int `yaml:"heartbeat_interval"` // 心跳间隔（秒）
	HeartbeatTimeout  int `yaml:"heartbeat_timeout"`  // 心跳超时（秒）
	ReadBufferSize    int `yaml:"read_buffer_size"`
	WriteBufferSize   int `yaml:"write_buffer_size"`
}

// Message 消息配置
type Message struct {
	MsgExpire       int `yaml:"msg_expire"`        // 消息保存天数
	MaxChatHistory  int `yaml:"max_chat_history"` // 单个对话最大历史消息数
	MsgProcessedTTL int `yaml:"msg_processed_ttl"` // 已处理消息保存时间（分钟）
}

// Cache 缓存配置
type Cache struct {
	UserInfoExpire int `yaml:"user_info_expire"` // 用户信息缓存过期时间（秒）
}
