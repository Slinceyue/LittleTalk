package router

import (
	"LittleTalk/handler/room_handler"

	"github.com/gin-gonic/gin"
)

func RoomRouter(r *gin.RouterGroup) {
	roomHandler := room_handler.RoomHandler{}

	// 群聊基本操作
	r.GET("/rooms", roomHandler.GetUserRooms)           // 获取我的群列表
	r.POST("/room", roomHandler.CreateRoom)              // 创建群聊
	r.GET("/room/:room_id", roomHandler.GetRoomInfo)    // 获取群信息
	r.GET("/room/:room_id/members", roomHandler.GetRoomMembers) // 获取群成员

	// 群聊操作
	r.POST("/room/join", roomHandler.JoinRoom)           // 加入群聊
	r.POST("/room/quit", roomHandler.QuitRoom)          // 退出群聊
	r.POST("/room/dismiss", roomHandler.DismissRoom)    // 解散群聊
	r.PUT("/room", roomHandler.UpdateRoom)              // 更新群信息

	// 群成员管理
	r.POST("/room/admin", roomHandler.SetAdmin)         // 设置管理员
	r.POST("/room/kick", roomHandler.KickMember)        // 踢出成员
	r.POST("/room/transfer", roomHandler.TransferOwner) // 转让群主
	r.POST("/room/invite", roomHandler.InviteMembers)   // 邀请成员入群

	// 搜索
	r.GET("/rooms/search", roomHandler.SearchRooms)    // 搜索群聊
}
