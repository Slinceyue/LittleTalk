package friend_handler

import (
	"LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/service"

	"github.com/gin-gonic/gin"
)

type FriendHandler struct{}

func (FriendHandler) GetFriendListHandler(c *gin.Context) {
	userID, _ := c.Get("id")
	friendList, err := service.GetFriendList(c.Request.Context(), userID.(uint))
	if err != nil {
		response.FailWithMsg(c, enum.CodeWrongResCode, "查找失败")
		return
	}
	response.OKWithData(c, friendList)
	return
}
func (FriendHandler) FriendRequestHandler(c *gin.Context) {
	userID, _ := c.Get("id")
	var _request request.FriendRequest
	if err := c.ShouldBindJSON(&_request); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}
	err := service.FriendRequest(c.Request.Context(), _request, userID.(uint))
	if err != nil {
		response.FailWithMsg(c, enum.CodeWrongResCode, "发送失败")
		return
	}
	response.OK(c)
}
func (FriendHandler) GetFriendRequestHandler(c *gin.Context) {
	userID, _ := c.Get("id")
	friendRequestList, err := service.GetFriendRequest(c.Request.Context(), userID.(uint))
	if err != nil {
		response.FailWithMsg(c, enum.CodeWrongResCode, "查找失败")
		return
	}
	response.OKWithData(c, friendRequestList)
}
func (FriendHandler) OKFriendRequestHandler(c *gin.Context) {
	userID, _ := c.Get("id")
	var _request request.FriendRequestOK
	if err := c.ShouldBindJSON(&_request); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}
	err := service.OKFriendRequest(c.Request.Context(), _request, userID.(uint))
	if err != nil {
		response.FailWithMsg(c, enum.CodeWrongResCode, "操作失败")
		return
	}
	response.OK(c)
}
