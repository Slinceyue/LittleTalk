package router

import (
	"LittleTalk/handler"

	"github.com/gin-gonic/gin"
)

func MessageRouter(r *gin.RouterGroup) {
	r.GET("/ws", handler.Api.MessageHandler.WS)
	r.GET("/message/list", handler.Api.MessageHandler.GetMessageList)
	r.GET("/message/db-history", handler.Api.MessageHandler.GetDBChatHistory)
	r.GET("/message/history", handler.Api.MessageHandler.GetChatHistory)
	r.GET("/message/search", handler.Api.MessageHandler.SearchMessages)
	r.GET("/unreadcount", handler.Api.MessageHandler.GetUnreadCount)
}
