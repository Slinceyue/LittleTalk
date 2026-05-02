package router

import (
	"LittleTalk/global"
	"LittleTalk/handler"

	"github.com/gin-gonic/gin"
)

func Run() {
	gin.SetMode(global.Config.System.GinMode)
	r := gin.Default()
	r.Use(handler.Api.CorsHandler)
	nr := r.Group("/api")

	nr.Use(handler.Api.MiddleHandler.ParseTokenHandler)
	ToolsRouter(nr)
	UserRouter(nr)
	MessageRouter(nr)
	UserLogin_Creat(r)
	FriendRouter(nr)
	r.StaticFile("/", "./test.html")
	addr := global.Config.System.Addr()
	r.Run(addr)
}
