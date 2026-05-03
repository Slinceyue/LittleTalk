package router

import (
	"LittleTalk/handler"

	"github.com/gin-gonic/gin"
)

func UserRouter(r *gin.RouterGroup) {
	UserAvatar(r)

	r.GET("/selfuserinfo", handler.Api.UserHandler.SelfUserInfo)
	r.GET("/otheruserinfo", handler.Api.UserHandler.OtherUserInfo)
	r.GET("/usersinfo", handler.Api.UserHandler.GetUsersInfo)
	r.POST("/offline", handler.Api.UserHandler.Offline)
	r.POST("/updateuserinfo", handler.Api.UserHandler.UpdateUserInfo)
}

func UserAvatar(r *gin.RouterGroup) {
	r.Static("/static/avatar", "static/avatar")
	r.POST("/uploadavatar", handler.Api.UserHandler.AvatarUpload)
}
func UserLogin_Creat(r *gin.Engine) {
	r.POST("/login", handler.Api.LoginHandler.Login)
	r.POST("/creatuser", handler.Api.UserHandler.CreatUserHandler)
}
