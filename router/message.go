package router

import (
	"LittleTalk/handler"

	"github.com/gin-gonic/gin"
)

func MessageRouter(r *gin.RouterGroup) {
	r.GET("/ws", handler.Api.MessageHandler.WS)
}
