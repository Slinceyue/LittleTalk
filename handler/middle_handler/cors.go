package middle_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 跨域中间件 - 允许所有来源
func (MiddleHandler) CorsHandler(c *gin.Context) {
	origin := c.GetHeader("Origin")

	// 处理预检 OPTIONS
	if c.Request.Method == "OPTIONS" {
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	// 同源请求直接放行
	if origin == "" {
		c.Next()
		return
	}

	// 跨域请求：回显具体 Origin（AllowCredentials 需要具体 Origin，不能是 *）
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Next()
}
