# LittleTalk API 文档

## 基础信息

- **Base URL**: `http://localhost:8080`
- **认证方式**: JWT Token (除登录和注册外，所有API需要通过Cookie或Header携带Token)

---

## 1. 用户认证

### 1.1 用户登录
```
POST /login
Content-MsgType: application/json

Request Body:
{
    "username": "string",    // 用户名
    "password": "string"      // 密码
}

Response:
{
    "code": 200,
    "data": "jwt_token_string"
}
```

### 1.2 用户注册
```
POST /creatuser
Content-MsgType: application/json

Request Body:
{
    "username": "string",    // 用户名
    "password": "string",     // 密码
    "sex": 1,                 // 性别 (0: 女, 1: 男, 2: 未知)
    "birthday": "2000-01-01"  // 生日 (格式: YYYY-MM-DD)
}

Response:
{
    "code": 200,
    "data": "注册成功信息"
}
```

---

## 2. 用户信息

### 2.1 获取自身信息
```
GET /api/selfuserinfo
Authorization: Bearer <token>

Response:
{
    "code": 200,
    "data": {
        "avatar": "static/avatar/{userID}.jpg",
        "username": "string",
        "sex": 1,
        "intro": "string",
        "birthday": "2000-01-01"
    }
}
```

### 2.2 获取其他用户信息
```
GET /api/otheruserinfo?userID={userID}
Authorization: Bearer <token>

Query Parameters:
- userID: uint, 目标用户ID

Response:
{
    "code": 200,
    "data": {
        "avatar": "static/avatar/{userID}.jpg",
        "username": "string",
        "sex": 1,
        "intro": "string",
        "birthday": "2000-01-01"
    }
}
```

### 2.3 上传头像
```
POST /api/uploadavatar
Authorization: Bearer <token>
Content-MsgType: multipart/form-data

Form Data:
- file: binary, JPG格式图片文件 (最大40MB)

Response:
{
    "code": 200,
    "data": "上传成功信息"
}
```

### 2.4 头像文件访问
```
GET /api/static/avatar/{userID}.jpg
```

---

## 3. 好友管理

### 3.1 获取好友列表
```
GET /api/friendlist
Authorization: Bearer <token>

Response:
{
    "code": 200,
    "data": [1, 2, 3]  // 好友用户ID数组
}
```

### 3.2 发送好友请求
```
POST /api/newfriendreq
Authorization: Bearer <token>
Content-MsgType: application/json

Request Body:
{
    "friendID": 2  // 目标用户ID
}

Response:
{
    "code": 200,
    "data": "发送成功信息"
}
```

### 3.3 获取好友请求列表
```
GET /api/friendreqlist
Authorization: Bearer <token>

Response:
{
    "code": 200,
    "data": [1, 2, 3]  // 发送好友请求的用户ID数组
}
```

### 3.4 接受好友请求
```
POST /api/okwithfriendreq
Authorization: Bearer <token>
Content-MsgType: application/json

Request Body:
{
    "fromID": 1  // 请求方用户ID
}

Response:
{
    "code": 200,
    "data": "添加成功信息"
}
```

---

## 4. 消息通信

### 4.1 WebSocket连接
```
GET /api/ws?token={jwt_token}

Header:
Authorization: Bearer <token>

WebSocket URL: ws://localhost:8080/api/ws?token=<token>

发送消息格式:
{
    "toID": 2,           // 接收方用户ID
    "message": "string"  // 消息内容
}

接收消息格式:
{
    "fromID": 1,         // 发送方用户ID
    "message": "string"  // 消息内容
}
```

---

## 5. 文件管理

### 5.1 上传文件
```
POST /api/uploadfile
Authorization: Bearer <token>
Content-MsgType: multipart/form-data

Form Data:
- file: binary, 文件 (最大40MB)
- fileType: int, 文件类型 (1: 图片, 2: 视频, 3: 音频, 4: 文档, 5: 其他)

Response:
{
    "code": 200,
    "data": "unique_filename"
}
```

### 5.2 下载文件
```
GET /api/downloadfile?fileType={type}&fromID={userID}&fileName={filename}
Authorization: Bearer <token>

Query Parameters:
- fileType: int, 文件类型
- fromID: uint, 上传用户ID
- fileName: string, 文件名

Response: 文件二进制流
```

---

## 6. 测试页面

### 6.1 API测试页面
```
GET /
```
访问 `http://localhost:8080/` 获取Web测试页面

---

## 响应码说明

| 响应码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权/Token无效 |
| 500 | 服务器内部错误 |

---

## 注意事项

1. **Token传递**: 除 `/login` 和 `/creatuser` 外，所有API需要在请求头中携带有效的JWT Token
2. **文件格式**: 头像仅支持JPG格式，文件上传限制40MB
3. **WebSocket**: 需要在URL参数中携带token进行连接认证
4. **Gin模式**: 当前为release模式，生产环境建议保持
