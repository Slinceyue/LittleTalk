package middle_handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// 跨域中间件
func (MiddleHandler) CorsHandler(c *gin.Context) {
	origin := c.GetHeader("Origin")

	// 处理预检 OPTIONS
	if c.Request.Method == "OPTIONS" {
		allowOrigin := ""
		if strings.HasPrefix(origin, "http://192.168.") ||
			strings.HasPrefix(origin, "http://10.") ||
			strings.HasPrefix(origin, "http://172.") {
			allowOrigin = origin
		} else if origin == "http://localhost:8080" ||
			origin == "http://localhost:8081" ||
			origin == "http://127.0.0.1:8080" ||
			origin == "http://127.0.0.1:8081" {
			allowOrigin = origin
		}

		if allowOrigin == "" && origin != "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
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

	// 跨域请求：检查是否在白名单
	allowOrigin := ""
	if strings.HasPrefix(origin, "http://192.168.") ||
		strings.HasPrefix(origin, "http://10.") ||
		strings.HasPrefix(origin, "http://172.") {
		allowOrigin = origin
	} else if origin == "http://localhost:8080" ||
		origin == "http://localhost:8081" ||
		origin == "http://127.0.0.1:8080" ||
		origin == "http://127.0.0.1:8081" {
		allowOrigin = origin
	}

	if allowOrigin == "" {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Header("Access-Control-Allow-Origin", allowOrigin)
	c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Next()
}
