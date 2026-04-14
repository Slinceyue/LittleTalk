package router

import (
	"github.com/gin-gonic/gin"
)

func ToolsRouter(r *gin.RouterGroup) {
	FileRouter(r)
}
func FileRouter(r *gin.RouterGroup) {
	r.POST("/uploadfile", func(c *gin.Context) {})
}
