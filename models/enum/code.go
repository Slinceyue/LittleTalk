package enum

import "fmt"

/**
 * 响应码设计原则：
 * - 0      = 成功
 * - 1xxxx  = 通用错误
 * - 2xxxx  = 用户模块
 * - 3xxxx  = 好友模块
 * - 4xxxx  = 消息模块
 * - 5xxxx  = 文件模块
 * - 6xxxx  = WebSocket模块
 */

// ResCode 响应码类型
type ResCode int

// Error 实现 error 接口，使 ResCode 可以作为 error 返回
func (c ResCode) Error() string {
	return c.String()
}

const (
	// ====== 成功 ======
	CodeSuccess ResCode = 0

	// ====== 通用错误 (1001-1999) ======
	CodeInvalidParam    ResCode = 1001 // 参数错误
	CodeUnauthorized    ResCode = 1002 // 未登录 / Token无效
	CodeForbidden       ResCode = 1003 // 无权限访问
	CodeNotFound        ResCode = 1004 // 资源不存在
	CodeTooManyRequests ResCode = 1005 // 请求过于频繁，请稍后重试
	CodeServerError     ResCode = 1500 // 服务器内部错误

	// ====== 用户模块 (2001-2999) ======
	CodeUserNotFound      ResCode = 2001 // 用户不存在
	CodeUserAlreadyExist  ResCode = 2002 // 用户已存在
	CodePasswordWrong     ResCode = 2003 // 密码错误
	CodePasswordWeak      ResCode = 2004 // 密码强度不足（至少6位）
	CodeInvalidUsername   ResCode = 2005 // 用户名格式错误（仅支持中英文、数字、下划线，2-20位）
	CodeInvalidPassword   ResCode = 2006 // 密码格式错误（至少6位）
	CodeUserCreateFailed   ResCode = 2100 // 用户创建失败
	CodeUserUpdateFailed   ResCode = 2101 // 用户信息更新失败

	// ====== 好友模块 (3001-3999) ======
	CodeFriendAlreadyExist     ResCode = 3001 // 已是好友，无需重复添加
	CodeFriendRequestExist     ResCode = 3002 // 好友申请已发送，请勿重复提交
	CodeFriendRequestNotFound  ResCode = 3003 // 好友申请不存在
	CodeFriendRequestExpired   ResCode = 3004 // 好友申请已过期
	CodeFriendRequestRejected  ResCode = 3005 // 好友申请已被拒绝
	CodeFriendSelfRequest      ResCode = 3006 // 不能添加自己为好友
	CodeFriendNotFound        ResCode = 3007 // 好友不存在
	CodeFriendRemoveFailed     ResCode = 3100 // 删除好友失败

	// ====== 消息模块 (4001-4999) ======
	CodeMessageEmpty        ResCode = 4001 // 消息内容不能为空
	CodeMessageTooLong      ResCode = 4002 // 消息内容过长（最多1000字符）
	CodeMessageSendFail     ResCode = 4003 // 消息发送失败，请重试
	CodeMessageNotFound     ResCode = 4004 // 消息不存在
	CodeMessageRecallFail   ResCode = 4005 // 消息撤回失败
	CodeMessageRecallExpire ResCode = 4006 // 超过撤回时限（2分钟内）

	// ====== 文件模块 (5001-5999) ======
	CodeFileLoadFail    ResCode = 5001 // 文件加载失败
	CodeFileUploadFail  ResCode = 5002 // 文件上传失败
	CodeFileTypeWrong   ResCode = 5003 // 文件类型不支持
	CodeFileTooLarge    ResCode = 5004 // 文件大小超出限制
	CodeFileEmpty       ResCode = 5005 // 文件内容为空
	CodeFileNotFound    ResCode = 5006 // 文件不存在
	CodeFileDataError   ResCode = 5007 // 文件数据损坏

	// ====== WebSocket模块 (6001-6999) ======
	CodeWsConnectFail      ResCode = 6001 // WebSocket连接失败
	CodeWsSendFail         ResCode = 6002 // WebSocket消息发送失败
	CodeWsInvalidMessage   ResCode = 6003 // 无效的WebSocket消息格式
	CodeWsHeartbeatTimeout ResCode = 6004 // 连接已断开（心跳超时）
	CodeWsAlreadyConnected ResCode = 6005 // WebSocket已连接
)

// codeMsgMap 响应码到消息的映射
var codeMsgMap = map[ResCode]string{
	CodeSuccess:             "操作成功",
	CodeInvalidParam:        "参数错误",
	CodeUnauthorized:        "未登录或登录已过期，请重新登录",
	CodeForbidden:           "无权限访问该资源",
	CodeNotFound:            "请求的资源不存在",
	CodeTooManyRequests:     "请求过于频繁，请稍后重试",
	CodeServerError:         "服务器开小差了，请稍后重试",

	CodeUserNotFound:        "用户不存在",
	CodeUserAlreadyExist:    "用户名已存在",
	CodePasswordWrong:       "密码错误",
	CodePasswordWeak:        "密码强度不足，请使用至少6位字符",
	CodeInvalidUsername:     "用户名格式错误，仅支持中英文、数字、下划线，2-20位",
	CodeInvalidPassword:     "密码格式错误，请使用至少6位字符",
	CodeUserCreateFailed:    "用户创建失败，请稍后重试",
	CodeUserUpdateFailed:    "用户信息更新失败，请稍后重试",

	CodeFriendAlreadyExist:     "已是好友，无需重复添加",
	CodeFriendRequestExist:     "好友申请已发送，请勿重复提交",
	CodeFriendRequestNotFound:  "好友申请不存在或已处理",
	CodeFriendRequestExpired:   "好友申请已过期",
	CodeFriendRequestRejected:  "好友申请已被拒绝",
	CodeFriendSelfRequest:      "不能添加自己为好友",
	CodeFriendNotFound:        "好友不存在",
	CodeFriendRemoveFailed:     "删除好友失败，请稍后重试",

	CodeMessageEmpty:        "消息内容不能为空",
	CodeMessageTooLong:      "消息内容过长，最多1000个字符",
	CodeMessageSendFail:     "消息发送失败，请重试",
	CodeMessageNotFound:     "消息不存在",
	CodeMessageRecallFail:   "消息撤回失败",
	CodeMessageRecallExpire: "超过撤回时限，2分钟内的消息才可撤回",

	CodeFileLoadFail:    "文件加载失败，请稍后重试",
	CodeFileUploadFail:  "文件上传失败，请稍后重试",
	CodeFileTypeWrong:   "不支持的文件类型",
	CodeFileTooLarge:    "文件大小超出限制",
	CodeFileEmpty:       "文件内容为空",
	CodeFileNotFound:    "文件不存在",
	CodeFileDataError:   "文件数据损坏",

	CodeWsConnectFail:      "WebSocket连接失败，请检查网络",
	CodeWsSendFail:         "消息发送失败，请稍后重试",
	CodeWsInvalidMessage:   "消息格式错误",
	CodeWsHeartbeatTimeout: "连接已断开，请刷新页面重试",
	CodeWsAlreadyConnected: "WebSocket已连接",
}

// String 获取状态码对应的中文提示信息
func (c ResCode) String() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		return "未知错误"
	}
	return msg
}

// Int 转换为 int 类型
func (c ResCode) Int() int {
	return int(c)
}

// ToMap 转换为 map 结构（用于 JSON 响应）
func (c ResCode) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"code":    c.Int(),
		"message": c.String(),
	}
}

// WrapError 包装错误信息（用于追加具体错误描述）
func (c ResCode) WrapError(err error) string {
	if err == nil {
		return c.String()
	}
	return fmt.Sprintf("%s: %v", c.String(), err)
}
