package room_handler

import (
	"LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/service"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseUintParam 解析路径参数为uint
func parseUintParam(param string, result *uint) (bool, error) {
	val, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return false, err
	}
	*result = uint(val)
	return true, nil
}

type RoomHandler struct{}

// CreateRoom 创建群聊
func (RoomHandler) CreateRoom(c *gin.Context) {
	var req request.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}

	userID := c.GetUint("id")
	room, err := service.CreateRoom(c.Request.Context(), req.Name, userID)
	if err != nil {
		if re, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, re)
		} else {
			response.FailWithCode(c, enum.CodeServerError)
		}
		return
	}

	response.OKWithData(c, room)
}

// GetUserRooms 获取用户的群列表
func (RoomHandler) GetUserRooms(c *gin.Context) {
	userID := c.GetUint("id")
	rooms, err := service.GetUserRooms(c.Request.Context(), userID)
	if err != nil {
		response.FailWithCode(c, enum.CodeServerError)
		return
	}

	response.OKWithData(c, rooms)
}

// GetRoomInfo 获取群信息
func (RoomHandler) GetRoomInfo(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	if roomIDStr == "" {
		response.FailWithMsg(c, enum.CodeInvalidParam, "缺少room_id参数")
		return
	}
	var roomID uint
	if _, err := parseUintParam(roomIDStr, &roomID); err != nil || roomID == 0 {
		response.FailWithMsg(c, enum.CodeInvalidParam, "无效的room_id")
		return
	}

	room, err := service.GetRoomInfo(c.Request.Context(), roomID)
	if err != nil {
		if re, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, re)
		} else {
			response.FailWithCode(c, enum.CodeServerError)
		}
		return
	}

	response.OKWithData(c, room)
}

// GetRoomMembers 获取群成员列表
func (RoomHandler) GetRoomMembers(c *gin.Context) {
	roomIDStr := c.Param("room_id")
	if roomIDStr == "" {
		response.FailWithMsg(c, enum.CodeInvalidParam, "缺少room_id参数")
		return
	}
	var roomID uint
	if _, err := parseUintParam(roomIDStr, &roomID); err != nil || roomID == 0 {
		response.FailWithMsg(c, enum.CodeInvalidParam, "无效的room_id")
		return
	}

	members, err := service.GetRoomMembers(c.Request.Context(), roomID)
	if err != nil {
		response.FailWithCode(c, enum.CodeServerError)
		return
	}

	response.OKWithData(c, members)
}

// JoinRoom 加入群聊
func (RoomHandler) JoinRoom(c *gin.Context) {
	var req request.JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}

	userID := c.GetUint("id")
	err := service.JoinRoom(c.Request.Context(), req.RoomID, userID)
	if err != nil {
		if re, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, re)
		} else {
			response.FailWithCode(c, enum.CodeServerError)
		}
		return
	}

	response.OK(c)
}

// QuitRoom 退出群聊
func (RoomHandler) QuitRoom(c *gin.Context) {
	var req request.JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}

	userID := c.GetUint("id")
	err := service.QuitRoom(c.Request.Context(), req.RoomID, userID)
	if err != nil {
		if re, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, re)
		} else {
			response.FailWithCode(c, enum.CodeServerError)
		}
		return
	}

	response.OK(c)
}

// DismissRoom 解散群聊
func (RoomHandler) DismissRoom(c *gin.Context) {
	var req request.JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}

	userID := c.GetUint("id")
	err := service.DismissRoom(c.Request.Context(), req.RoomID, userID)
	if err != nil {
		if re, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, re)
		} else {
			response.FailWithCode(c, enum.CodeServerError)
		}
		return
	}

	response.OK(c)
}

// UpdateRoom 更新群信息
func (RoomHandler) UpdateRoom(c *gin.Context) {
	var req request.UpdateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}

	userID := c.GetUint("id")
	err := service.UpdateRoom(c.Request.Context(), req.RoomID, req.Name, req.Avatar, req.Intro, userID)
	if err != nil {
		if re, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, re)
		} else {
			response.FailWithCode(c, enum.CodeServerError)
		}
		return
	}

	response.OK(c)
}

// SetAdmin 设置/取消管理员
func (RoomHandler) SetAdmin(c *gin.Context) {
	var req request.SetAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}

	userID := c.GetUint("id")
	err := service.SetRoomAdmin(c.Request.Context(), req.RoomID, req.TargetUserID, userID, req.IsAdmin)
	if err != nil {
		if re, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, re)
		} else {
			response.FailWithCode(c, enum.CodeServerError)
		}
		return
	}

	response.OK(c)
}

// KickMember 踢出成员
func (RoomHandler) KickMember(c *gin.Context) {
	var req request.KickMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}

	userID := c.GetUint("id")
	err := service.KickRoomMember(c.Request.Context(), req.RoomID, req.TargetUserID, userID)
	if err != nil {
		if re, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, re)
		} else {
			response.FailWithCode(c, enum.CodeServerError)
		}
		return
	}

	response.OK(c)
}

// TransferOwner 转让群主
func (RoomHandler) TransferOwner(c *gin.Context) {
	var req request.TransferOwnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}

	userID := c.GetUint("id")
	err := service.TransferRoomOwner(c.Request.Context(), req.RoomID, req.TargetUserID, userID)
	if err != nil {
		if re, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, re)
		} else {
			response.FailWithCode(c, enum.CodeServerError)
		}
		return
	}

	response.OK(c)
}

// SearchRooms 搜索群聊
func (RoomHandler) SearchRooms(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.FailWithMsg(c, enum.CodeInvalidParam, "请输入搜索关键词")
		return
	}

	rooms, err := service.SearchRooms(c.Request.Context(), keyword)
	if err != nil {
		log.Printf("[SearchRooms] error: %v", err)
		response.FailWithCode(c, enum.CodeServerError)
		return
	}

	response.OKWithData(c, rooms)
}

// InviteMembers 邀请成员入群
func (RoomHandler) InviteMembers(c *gin.Context) {
	var req request.InviteMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "参数错误")
		return
	}

	if len(req.TargetUserIDs) == 0 {
		response.FailWithMsg(c, enum.CodeInvalidParam, "请选择要邀请的好友")
		return
	}

	userID := c.GetUint("id")
	addedCount, err := service.InviteMembers(c.Request.Context(), req.RoomID, req.TargetUserIDs, userID)
	if err != nil {
		if re, ok := err.(enum.ResCode); ok {
			response.FailWithCode(c, re)
		} else {
			response.FailWithCode(c, enum.CodeServerError)
		}
		return
	}

	response.OKWithData(c, gin.H{
		"added_count": addedCount,
		"message":     fmt.Sprintf("成功邀请 %d 位好友入群", addedCount),
	})
}
