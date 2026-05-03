package user_handler

import (
	"LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/cache"
	"LittleTalk/models/enum"
	"LittleTalk/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type LoginHandler struct {
}

type LoginResponse struct {
	Token  string `json:"token"`
	UserID uint   `json:"user_id"`
}

func (LoginHandler) Login(c *gin.Context) {
	var loginRequest request.LoginRequest
	err := c.ShouldBindJSON(&loginRequest)
	if err != nil {
		logrus.Warn("登录参数错误: ", err)
		response.FailWithCode(c, enum.CodeInvalidParam)
		return
	}
	logrus.Info("Login attempt for username: ", loginRequest.Username, " userID: ", loginRequest.UserID)
	ctx := c.Request.Context()
	result, err := service.LoginUser(ctx, loginRequest)
	if err != nil {
		logrus.Error("登录失败 for ", loginRequest.Username, ": ", err)
		response.FailWithError(c, enum.CodeUnauthorized, err)
		return
	}
	// 设置cookie (HttpOnly=false 允许前端JavaScript访问)
	c.SetCookie("token", result.Token, 86400*7, "/", "", false, false)
	// 设置用户在线状态
	if err := cache.SetUserOnlineStatus(ctx, result.UserID); err != nil {
		logrus.Warn("设置用户在线状态失败: ", err)
	}
	logrus.Info("用户 ", result.UserID, " 登录成功")
	response.OKWithData(c, LoginResponse{
		Token:  result.Token,
		UserID: result.UserID,
	})
}
