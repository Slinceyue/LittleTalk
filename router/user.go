package router

import (
	"LittleTalk/handler"

	"github.com/gin-gonic/gin"
)

func UserRouter(r *gin.RouterGroup) {
	UserAvatar(r)
}

func UserAvatar(r *gin.RouterGroup) {
	r.Static("/static/avatar", "static/avatar")
	r.POST("/uploadavatar", handler.Api.UserHandler.AvatarUpload)
}
