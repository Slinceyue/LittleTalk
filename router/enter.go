package router

import (
	"LittleTalk/global"
	"LittleTalk/handler"

	"github.com/gin-gonic/gin"
)

func Run() {
	gin.SetMode(global.Config.System.GinMode)
	r := gin.Default()
	nr := r.Group("/api")

	nr.Use(handler.Api.MiddleHandler.ParseTokenHandler)
	ToolsRouter(nr)
	UserRouter(nr)
	addr := global.Config.System.Addr()
	r.Run(addr)
}
