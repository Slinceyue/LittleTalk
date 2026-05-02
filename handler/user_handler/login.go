package user_handler

import (
	"LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type LoginHandler struct {
}

func (LoginHandler) Login(c *gin.Context) {
	var loginRequest request.LoginRequest
	err := c.ShouldBindJSON(&loginRequest)
	if err != nil {
		response.FailWithCode(c, enum.CodeInvalidParam)
		return
	}
	logrus.Info("Login attempt for username: ", loginRequest.Username, " userID: ", loginRequest.UserID)
	ctx := c.Request.Context()
	token, err := service.LoginUser(ctx, loginRequest)
	if err != nil {
		response.FailWithError(c, enum.CodeUnauthorized, err)
		return
	}
	// 设置cookie
	c.SetCookie("token", token, 86400*7, "/", "", false, true)
	response.OKWithData(c, response.Response{
		Code:    enum.CodeSuccess.Int(),
		Message: "登录成功",
		Data:    token,
	})
}
