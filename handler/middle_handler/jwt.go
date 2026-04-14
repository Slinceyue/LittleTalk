package middle_handler

import (
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/utils/jwts"

	"github.com/gin-gonic/gin"
)

func (MiddleHandler) ParseTokenHandler(c *gin.Context) {
	// 从Cookie获取token
	tokenStr, err := c.Cookie("token")
	if err != nil {
		response.FailWithCode(c, enum.CodeUnauthorized)
		c.Abort()
		return
	}

	// 解析并验证JWT
	claims, err := jwts.ParseToken(tokenStr)
	if err != nil {
		response.FailWithError(c, enum.CodeUnauthorized, err)
		c.Abort()
		return
	}

	// 存入上下文，后续路由可直接获取
	c.Set("claims", claims)
	c.Next()
}
