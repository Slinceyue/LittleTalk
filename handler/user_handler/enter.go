package user_handler

import (
	request2 "LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/service"
	"strconv"

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
