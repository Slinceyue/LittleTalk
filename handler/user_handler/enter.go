package user_handler

import (
	request2 "LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/service"
	"LittleTalk/utils/ws"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{}

func (UserHandler) CreatUserHandler(c *gin.Context) {
	var request request2.NewUserRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}
	err = service.CreatUser(c.Request.Context(), request)
	if err != nil {
		response.FailWithMsg(c, enum.CodeUserCreateFailed, "创建用户失败")
		return
	}
	response.OK(c)
}
func (UserHandler) SelfUserInfo(c *gin.Context) {
	id, _ := c.Get("id")
	UserInf, _ := service.GetUser(c.Request.Context(), id.(uint))
	response.OKWithData(c, UserInf)
}

func (UserHandler) OtherUserInfo(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		id = c.Query("user_id")
	}
	num64, err := strconv.ParseUint(id, 10, 0)
	if err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "非法ID")
		return
	}
	UserInf, err := service.GetOtherUser(c.Request.Context(), uint(num64))
	if err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "用户不存在")
		return
	}
	response.OKWithData(c, UserInf)
}

// GetUsersInfo 批量获取用户信息（返回id、username、avatar）
func (UserHandler) GetUsersInfo(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		response.FailWithMsg(c, enum.CodeInvalidParam, "缺少ids参数")
		return
	}

	// 解析ID列表
	idStrs := strings.Split(idsStr, ",")
	var ids []uint
	for _, s := range idStrs {
		id, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
		if err == nil {
			ids = append(ids, uint(id))
		}
	}

	if len(ids) == 0 {
		response.OKWithData(c, []interface{}{})
		return
	}

	users, err := service.GetUserInfosByIDs(c.Request.Context(), ids)
	if err != nil {
		response.FailWithMsg(c, enum.CodeServerError, "获取用户信息失败")
		return
	}

	response.OKWithData(c, users)
}

// Offline 用户主动下线（前端页面关闭时调用）
// 这是一个辅助优化，用于加速正常关闭时的下线
// 即使前端未调用，WS 断开时后端也会自动清理
func (UserHandler) Offline(c *gin.Context) {
	id, exists := c.Get("id")
	if !exists {
		response.FailWithMsg(c, enum.CodeUnauthorized, "未登录")
		return
	}
	userID := id.(uint)

	// 清理 WS 连接（内部会删除 Redis 在线状态）
	ws.ConnManager.Delete(userID)

	response.OK(c)
}

// UpdateUserInfo 更新用户个人信息
func (UserHandler) UpdateUserInfo(c *gin.Context) {
	id, exists := c.Get("id")
	if !exists {
		response.FailWithMsg(c, enum.CodeUnauthorized, "未登录")
		return
	}
	userID := id.(uint)

	var req request2.UserUpdateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}

	// 设置用户ID
	req.UserID = userID

	code := service.UserInfoUpdate(c.Request.Context(), userID, req)
	if code != enum.CodeSuccess {
		// 处理错误码
		switch code {
		case enum.CodeUserNotFound:
			response.FailWithMsg(c, enum.CodeUserNotFound, "用户不存在")
		case enum.CodeUserAlreadyExist:
			response.FailWithMsg(c, enum.CodeUserAlreadyExist, "用户名已被占用")
		default:
			response.FailWithMsg(c, enum.CodeServerError, "更新失败")
		}
		return
	}

	response.OK(c)
}
