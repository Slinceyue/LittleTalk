

# LittleTalk - 即时通讯系统

基于 Go + Gin + GORM 构建的轻量级即时通讯后端服务，支持用户管理、好友系统、实时消息、WebSocket 实时通信等功能。前端采用原生 HTML/CSS/JavaScript 实现，提供友好的用户界面。

## 📋 目录

- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [API 接口](#api-接口)
- [数据库设计](#数据库设计)
- [WebSocket 实时通信](#websocket-实时通信)
- [开发规范](#开发规范)

## 🛠 技术栈

### 后端核心框架
- **Go**: 1.21+
- **Web 框架**: Gin v1.12.0
- **ORM**: GORM v1.31.1 + MySQL Driver
- **数据库**: MySQL (支持主从读写分离)
- **缓存**: Redis

### 前端
- **核心**: 原生 HTML5 + CSS3 + JavaScript (ES6+)
- **UI**: 自定义组件，支持亮色/暗色主题切换
- **通信**: Axios (HTTP) + 原生 WebSocket

### 关键依赖
- **JWT 认证**: golang-jwt/jwt/v5
- **参数验证**: go-playground/validator/v10
- **日志系统**: logrus
- **配置文件**: goccy/go-yaml
- **IP 定位**: ip2region
- **WebSocket**: gorilla/websocket

## 📁 项目结构

```

LittleTalk/
├── api/                    # API 层：请求/响应数据结构定义
│   ├── _request/           # 请求参数结构体
│   │   ├── enter.go       # 请求模块入口
│   │   └── user.go        # 用户相关请求
│   └── resbonse/          # 响应数据结构体
│       ├── enter.go       # 响应模块入口
│       ├── fail.go        # 失败响应
│       └── success.go     # 成功响应
│
├── conf/                  # 配置结构体定义
│   ├── enter.go           # 配置模块入口
│   ├── conf_db.go         # 数据库配置
│   ├── conf_jwt.go        # JWT 配置
│   ├── conf_logrous.go    # 日志配置
│   └── conf_system.go     # 系统配置
│
├── core/                  # 核心初始化模块
│   ├── init_conf.go       # 配置文件读取
│   ├── init_db.go         # 数据库初始化
│   ├── init_ip_db.go      # IP 数据库初始化
│   ├── init_logrus.go     # 日志系统初始化
│   └── init_redis.go      # Redis 缓存初始化
│
├── dao/                   # 数据访问层（Data Access Object）
│   ├── enter.go           # DAO 模块入口
│   ├── user.go            # 用户数据操作
│   ├── friend.go          # 好友数据操作
│   └── messatge.go        # 消息数据操作
│
├── flags/                 # 命令行参数解析
│   ├── enter.go           # 命令行模块入口
│   └── flag_db.go         # 数据库相关命令行参数
│
├── global/                # 全局变量
│   └── enter.go           # 全局配置、数据库连接等
│
├── handler/               # 处理器层（HTTP 请求处理）
│   ├── enter.go           # Handler 模块入口
│   ├── login_handler/     # 登录注册处理器
│   │   └── enter.go
│   ├── user_handler/      # 用户信息处理器
│   │   ├── enter.go
│   │   └── user_info_api.go
│   ├── friend_handler/    # 好友管理处理器
│   │   └── enter.go
│   ├── message_handler/   # 消息处理
│   │   └── enter.go
│   └── tools_handler/     # 工具类处理器
│
├── models/                # 数据模型层
│   ├── enter.go           # 基础模型（含 gorm.Model）
│   ├── user_model.go      # 用户模型
│   ├── friend_model.go    # 好友关系模型
│   ├── friend_request_model.go  # 好友申请模型
│   ├── message_model.go   # 消息模型
│   ├── room_model.go      # 聊天室模型
│   └── enum/              # 枚举常量定义
│       ├── enter.go
│       ├── code.go        # 状态码
│       ├── user.go        # 用户相关枚举（性别、角色、在线状态）
│       ├── friend.go      # 好友状态
│       ├── friend_request.go  # 好友申请状态
│       ├── file.go        # 文件类型
│       └── message.go     # 消息状态
│
├── router/                # 路由层
│   ├── enter.go           # 路由注册入口
│   ├── user.go            # 用户相关路由
│   └── tools.go           # 工具类路由（文件上传等）
│
├── service/               # 业务逻辑层
│   ├── enter.go           # Service 模块入口
│   ├── user.go            # 用户业务逻辑
│   ├── friend.go          # 好友业务逻辑
│   ├── message.go         # 消息业务逻辑
│   └── file.go            # 文件处理业务
│
├── utils/                 # 工具包
│   ├── jwts/              # JWT 工具
│   │   └── enter.go
│   ├── ip/                # IP 定位工具
│   │   └── enter.go
│   └── validate/          # 参数验证工具
│       └── enter.go
│
├── web/                   # 前端资源目录
│   ├── index.html         # 主页面
│   ├── css/               # 样式文件
│   │   └── styles.css     # 全局样式
│   └── js/                # JavaScript 文件
│       └── app.js         # 应用逻辑
│
├── cache/                 # 缓存层
│   ├── enter.go           # 缓存模块入口
│   ├── user_cache.go      # 用户缓存
│   ├── friend.go          # 好友缓存
│   └── message.go         # 消息缓存
│
├── static/                # 静态资源
│   └── files/             # 上传文件存储
│       ├── avatar/        # 头像存储
│       └── file/          # 普通文件存储
│
├── logs/                  # 日志文件目录
│   └── {date}/            # 按日期分类
│       └── LittleTalk.log
│
├── main.go                # 程序入口
├── settings.yaml          # 应用配置文件
├── go.mod                 # Go 模块依赖
├── go.sum                 # 依赖校验文件
├── API_DOC.md             # API 接口文档（详细）
└── readme.md              # 项目说明文档
```
## 🚀 快速开始

### 环境要求

- Go >= 1.21
- MySQL >= 5.7
- Redis >= 6.0
- Git

### 安装步骤

1. **克隆项目**
```
bash
git clone <repository-url>
cd LittleTalk
```
2. **安装依赖**
```
bash
go mod download
```
3. **配置数据库**

创建数据库：
```
sql
CREATE DATABASE little_talk CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```
4. **修改配置文件**

编辑 `settings.yaml`，配置数据库连接等信息：
```yaml
system:
  ip: 0.0.0.0            # 绑定地址
  port: 8080             # 服务端口
  env: dev               # 运行环境：dev/prod
  gin_mode: debug        # Gin 运行模式：debug/release
  server_time_out: 2     # 服务器超时时间（秒）

log:
  app: LittleTalk        # 应用名称
  dir: logs              # 日志存储目录

jwt:
  expire: 3              # Token 过期时间（小时）
  secret: your-secret    # 签名密钥
  issuer: slince         # 签发者

redis:
  db: 0                  # Redis 数据库编号
  host: 127.0.0.1        # Redis 地址
  port: 6379             # Redis 端口
  password: xxx           # Redis 密码
  pool_size: 20          # 连接池大小
  online_expire: 50      # 在线状态过期时间（秒）

websocket:
  heartbeat_interval: 15 # 心跳间隔（秒）
  heartbeat_timeout: 45  # 心跳超时（秒）
  read_buffer_size: 1024 # 读缓冲区大小
  write_buffer_size: 1024# 写缓冲区大小

message:
  msg_expire: 7          # 消息保存天数
  max_chat_history: 100  # 单个对话最大历史消息数
  msg_processed_ttl: 5   # 已处理消息保存时间（分钟）

cache:
  user_info_expire: 3600 # 用户信息缓存过期时间（秒），默认1小时

db:                        # 从库配置（读操作）
  user: root
  password: your_password
  host: 127.0.0.1
  port: 3306
  db: little_talk
  debug: false
  source: mysql

db_master:                 # 主库配置（写操作）
  user: root
  password: your_password
  host: 127.0.0.1
  port: 3306
  db: little_talk
  debug: false
  source: mysql
```
5. **启动服务**
```
bash
go run main.go
```
服务将在 `http://localhost:8080` 启动。

## ⚙️ 配置说明

### 系统配置 (system)
- `ip`: 绑定地址（默认所有地址）
- `port`: 服务端口（默认 8080）
- `env`: 运行环境（dev/prod）
- `gin_mode`: Gin 框架模式（debug/release/test）
- `server_time_out`: 服务器超时时间（秒）

### 数据库配置 (db/db_master)
支持主从配置：
- `db`: 从库配置（读操作）
- `db_master`: 主库配置（写操作）

### JWT 配置 (jwt)
- `secret`: 签名密钥
- `expire`: Token 过期时间（**天**）
- `issuer`: 签发者

### 日志配置 (log)
- `app`: 应用名称
- `dir`: 日志存储目录

### Redis 配置 (redis)
- `host`: Redis 服务器地址
- `port`: Redis 端口
- `password`: Redis 密码
- `db`: 数据库编号
- `pool_size`: 连接池大小
- `online_expire`: 用户在线状态过期时间（**秒**）

### WebSocket 配置 (websocket)
- `heartbeat_interval`: 心跳间隔（秒）
- `heartbeat_timeout`: 心跳超时时间（秒）
- `read_buffer_size`: 读缓冲区大小
- `write_buffer_size`: 写缓冲区大小

### 消息配置 (message)
- `msg_expire`: 消息保存天数
- `max_chat_history`: 单个对话最大历史消息数
- `msg_processed_ttl`: 已处理消息保存时间（分钟）

### 缓存配置 (cache)
- `user_info_expire`: 用户信息缓存过期时间（秒）

## 📡 API 接口

> 详细的 API 文档请参考 [API_DOC.md](./API_DOC.md)

### 认证模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/login` | 用户登录 |
| POST | `/creatuser` | 用户注册 |

### 用户模块
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/selfuserinfo` | 获取自身信息 |
| GET | `/api/otheruserinfo` | 获取其他用户信息 |
| POST | `/api/uploadavatar` | 上传头像 |

### 好友模块
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/friendlist` | 获取好友列表 |
| POST | `/api/newfriendreq` | 发送好友请求 |
| GET | `/api/friendreqlist` | 获取好友请求列表 |
| POST | `/api/okwithfriendreq` | 处理好友请求（同意/拒绝） |
| DELETE | `/api/friend/:id` | 删除好友 |

### 消息模块
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/ws` | WebSocket 连接 |
| GET | `/api/messages` | 获取历史消息 |

### 文件模块
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/uploadfile` | 文件上传 |
| GET | `/api/downloadfile` | 文件下载 |

### 测试页面
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/` | API 测试页面 |

## 🗄️ 数据库设计

### 用户表 (users)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| username | varchar(10) | 用户名（唯一） |
| password | varchar(256) | 密码（加密存储） |
| sex | tinyint | 性别（0未知/1男/2女） |
| avatar | varchar(255) | 头像URL |
| intro | varchar(255) | 个人简介 |
| phone | varchar(16) | 手机号（唯一） |
| email | varchar(64) | 邮箱（唯一） |
| birthday | varchar(20) | 生日 |
| status | tinyint | 账号状态（1正常/2禁用） |
| role | tinyint | 角色（1普通/2管理员） |
| last_login | datetime | 最后登录时间 |
| ip | varchar(64) | 最后登录IP |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 删除时间（软删除） |

### 好友关系表 (friends)
| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | uint | 用户ID（联合主键） |
| friend_id | uint | 好友ID（联合主键） |
| remark | varchar(32) | 备注名 |
| status | tinyint | 好友状态 |

### 好友申请表 (friend_requests)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| from_user_id | uint | 申请人ID |
| to_user_id | uint | 被申请人ID |
| status | tinyint | 申请状态（0待处理/1同意/2拒绝） |
| created_at | datetime | 申请时间 |

### 消息表 (messages)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| from_id | uint | 发送者ID |
| to_id | uint | 接收者ID |
| content | varchar(1024) | 消息内容 |
| is_read | boolean | 是否已读 |
| created_at | datetime | 发送时间 |

## 🔌 WebSocket 实时通信

### 连接方式

```javascript
// 前端连接 WebSocket
const ws = new WebSocket('ws://localhost:8080/api/ws?token=' + jwtToken);
```

### 发送消息

```javascript
// 发送文本消息
ws.send(JSON.stringify({
    toID: 123,           // 接收方用户ID
    message: "你好！"     // 消息内容
}));
```

### 接收消息

```javascript
ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('收到消息:', data);
    // data.fromID   - 发送者用户ID
    // data.message  - 消息内容
    // data.time     - 发送时间
};
```

### 心跳机制

系统内置心跳检测：
- 客户端每 15 秒发送一次心跳
- 服务器 45 秒未收到心跳则断开连接
- 确保实时通信的稳定性

### 在线状态

用户上线/下线时，系统会自动更新 Redis 中的在线状态缓存，前端可通过轮询或 WebSocket 消息感知好友的在线状态变化。

## 🎨 前端说明

### 项目结构

```
web/
├── index.html    # 主页面
├── css/
│   └── styles.css   # 全局样式（支持暗色模式）
└── js/
    └── app.js       # 应用逻辑（Axios + WebSocket）
```

### 功能特性

- **用户认证**: 登录、注册、Token 自动管理
- **好友管理**: 好友列表、添加好友、好友请求处理
- **即时通讯**: 实时消息收发、消息气泡、已读未读状态
- **个人资料**: 头像上传、个人信息编辑
- **聊天背景**: 自定义聊天背景设置
- **主题切换**: 支持亮色/暗色主题
- **响应式设计**: 适配移动端和桌面端

### 使用方式

前端页面无需单独启动，直接访问后端服务：

```
http://{localhost}:{port}/web
```

后端服务会自动提供前端静态资源。

### 分层架构
项目采用经典的分层架构：

```

请求流程：
Router → Handler → Service → DAO → Database
↓           ↓
Request    Business Logic
↓
Response
```
- **Router**: 路由注册，定义 URL 映射
- **Handler**: 接收请求、参数验证、调用 Service、返回响应
- **Service**: 核心业务逻辑处理
- **DAO**: 数据库 CRUD 操作
- **Models**: 数据模型定义

### 代码规范

1. **命名规范**
   - 包名：小写，简短有意义
   - 文件名：小写，下划线分隔
   - 结构体：大驼峰（PascalCase）
   - 变量/函数：小驼峰（camelCase）

2. **错误处理**
   - 所有错误必须处理
   - 使用统一的响应格式

3. **日志记录**
   - 关键操作记录日志
   - 错误信息包含上下文

### 添加新功能流程

1. 在 `models/` 定义数据模型
2. 在 `dao/` 实现数据访问方法
3. 在 `service/` 编写业务逻辑
4. 在 `handler/` 创建 HTTP 处理器
5. 在 `router/` 注册路由
6. 在 `api/_request` 和 `api/response` 定义请求/响应结构

## 🔐 安全说明

- 密码使用 bcrypt 加密存储
- JWT Token 进行身份验证
- 参数验证防止 SQL 注入
- 敏感配置不提交到版本控制

## 📄 License

本项目仅供学习交流使用。

---

**注意**: 生产环境部署前，请务必修改默认配置，特别是 JWT secret 和数据库密码。
```