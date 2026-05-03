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

	// API路由组
	nr := r.Group("/api")
	nr.Use(handler.Api.MiddleHandler.ParseTokenHandler)
	ToolsRouter(nr)
	UserRouter(nr)
	MessageRouter(nr)
	FriendRouter(nr)
	UserLogin_Creat(r)

	// 静态文件服务（前端页面、头像等）
	r.Static("/web", "./web")
	r.Static("/static", "./static")

	// 前端页面入口（根路径）
	r.LoadHTMLFiles("./web/index.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "", nil)
	})

	addr := global.Config.System.Addr()
	r.Run(addr)
}
