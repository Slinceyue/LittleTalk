package router

import (
	"LittleTalk/handler"

	"github.com/gin-gonic/gin"
)

func FriendRouter(r *gin.RouterGroup) {
	r.GET("/friendlist", handler.Api.FriendHandler.GetFriendListHandler)
	r.POST("/newfriendreq", handler.Api.FriendHandler.FriendRequestHandler)
	r.GET("/friendreqlist", handler.Api.FriendHandler.GetFriendRequestHandler)
	r.POST("/okwithfriendreq", handler.Api.FriendHandler.OKFriendRequestHandler)
}
