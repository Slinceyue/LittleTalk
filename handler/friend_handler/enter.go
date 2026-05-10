package friend_handler

import (
	"LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/service"

	"github.com/gin-gonic/gin"
)

type FriendHandler struct{}

// GetFriendListHandler 获取好友列表
func (FriendHandler) GetFriendListHandler(c *gin.Context) {
	userID, _ := c.Get("id")
	friendList, err := service.GetFriendList(c.Request.Context(), userID.(uint))
	if err != nil {
		response.FailWithMsg(c, enum.CodeServerError, err.Error())
		return
	}
	response.OKWithData(c, friendList)
}

// FriendRequestHandler 发送好友请求
func (FriendHandler) FriendRequestHandler(c *gin.Context) {
	userID, _ := c.Get("id")
	var req request.FriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}
	err := service.FriendRequest(c.Request.Context(), req, userID.(uint))
	if err != nil {
		if code, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, code)
		} else {
			response.FailWithMsg(c, enum.CodeServerError, err.Error())
		}
		return
	}
	response.OK(c)
}

// GetFriendRequestHandler 获取好友请求列表
func (FriendHandler) GetFriendRequestHandler(c *gin.Context) {
	userID, _ := c.Get("id")
	friendRequestList, err := service.GetFriendRequest(c.Request.Context(), userID.(uint))
	if err != nil {
		response.FailWithMsg(c, enum.CodeServerError, err.Error())
		return
	}
	response.OKWithData(c, friendRequestList)
}

// OKFriendRequestHandler 同意好友请求
func (FriendHandler) OKFriendRequestHandler(c *gin.Context) {
	userID, _ := c.Get("id")
	var req request.FriendRequestOK
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}
	err := service.OKFriendRequest(c.Request.Context(), req, userID.(uint))
	if err != nil {
		if code, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, code)
		} else {
			response.FailWithMsg(c, enum.CodeServerError, err.Error())
		}
		return
	}
	response.OK(c)
}

// RejectFriendRequestHandler 拒绝好友请求
func (FriendHandler) RejectFriendRequestHandler(c *gin.Context) {
	userID, _ := c.Get("id")
	var req request.RejectFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}
	err := service.RejectFriendRequest(c.Request.Context(), req.FromID, userID.(uint))
	if err != nil {
		if code, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, code)
		} else {
			response.FailWithMsg(c, enum.CodeServerError, err.Error())
		}
		return
	}
	response.OK(c)
}

// DeleteFriendHandler 删除好友
func (FriendHandler) DeleteFriendHandler(c *gin.Context) {
	userID, _ := c.Get("id")
	var req request.DeleteFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}
	err := service.DeleteFriend(c.Request.Context(), req.FriendID, userID.(uint))
	if err != nil {
		if code, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, code)
		} else {
			response.FailWithMsg(c, enum.CodeServerError, err.Error())
		}
		return
	}
	response.OK(c)
}
