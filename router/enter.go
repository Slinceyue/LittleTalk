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
	RoomRouter(nr)
	UserLogin_Creat(r)

	// 静态文件服务（头像等）
	r.Static("/static", "./static")

	// 前端页面（Vue构建产物）
	r.Static("/assets", "./web/dist/assets")
	r.LoadHTMLFiles("./web/dist/index.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	addr := global.Config.System.Addr()
	r.Run(addr)
}
