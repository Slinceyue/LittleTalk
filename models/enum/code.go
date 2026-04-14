package enum

type ResCode int

// 用 iota 生成枚举值（分模块规划，避免混乱）
const (
	// 成功（0 开头）
	CodeSuccess      ResCode = 0
	CodeWrongResCode ResCode = 1

	// 通用错误（1 开头）
	CodeInvalidParam ResCode = 1001 // 参数错误
	CodeUnauthorized ResCode = 1002 // 未登录
	CodeForbidden    ResCode = 1003 // 无权限
	CodeServerError  ResCode = 1004 // 服务器内部错误

	// 用户模块（2 开头）
	CodeUserNotFound     ResCode = 2001 // 用户不存在
	CodeUserAlreadyExist ResCode = 2002 // 用户已存在
	CodePasswordWrong    ResCode = 2003 // 密码错误
	CodeUserCreateFailed ResCode = 2004 //用户创建失败

	// 好友模块（3 开头）
	CodeFriendAlreadyExist ResCode = 3001 // 已是好友
	CodeFriendRequestExist ResCode = 3002 // 申请已发送
	CodeFriendNotExist     ResCode = 3003 // 好友不存在

	// 消息模块（4 开头）
	CodeMessageSendFail ResCode = 4001 // 消息发送失败

	// 文件模块 (5 开头 )
	CodeFileLoadFail   ResCode = 5001 //文件加载失败
	CodeFileUploadFail ResCode = 5002 //文件上传失败
	CodeFileTypeWrong  ResCode = 5003 //文件类型错误
	CodeFileWrong      ResCode = 5004 //文件数据错误
)

var codeMsgMap = map[ResCode]string{
	CodeWrongResCode:       "错误",
	CodeSuccess:            "成功",
	CodeInvalidParam:       "参数错误",
	CodeUnauthorized:       "未登录",
	CodeForbidden:          "无权限",
	CodeServerError:        "服务器开小差了",
	CodeUserNotFound:       "用户不存在",
	CodeUserAlreadyExist:   "用户已存在",
	CodePasswordWrong:      "密码错误",
	CodeUserCreateFailed:   "用户创建失败",
	CodeFriendAlreadyExist: "已是好友",
	CodeFriendRequestExist: "申请已发送",
	CodeFriendNotExist:     "好友不存在",
	CodeMessageSendFail:    "消息发送失败",
	CodeFileLoadFail:       "文件加载失败",
	CodeFileUploadFail:     "文件上传失败",
	CodeFileTypeWrong:      "文件类型错误",
	CodeFileWrong:          "文件数据错误",
}

// Message 获取状态码对应的提示信息
func (c ResCode) Message() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		return codeMsgMap[CodeServerError]
	}
	return msg
}

// Int 转换为 int 类型
func (c ResCode) Int() int {
	return int(c)
}
