package router

import (
	"LittleTalk/handler"

	"github.com/gin-gonic/gin"
)

func ToolsRouter(r *gin.RouterGroup) {
	FileRouter(r)
}
func FileRouter(r *gin.RouterGroup) {
	r.POST("/uploadfile", handler.Api.ToolsHandler.FileUploadHandler)
	r.GET("/downloadfile", handler.Api.ToolsHandler.ReadFileHandler)
}
