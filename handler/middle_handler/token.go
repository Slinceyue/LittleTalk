package middle_handler

import (
	"LittleTalk/api/response"
	"LittleTalk/global"
	"LittleTalk/models/enum"
	"LittleTalk/utils/jwts"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

func (MiddleHandler) ParseTokenHandler(c *gin.Context) {
	// 放行OPTIONS预检请求
	if c.Request.Method == "OPTIONS" {
		c.Next()
		return
	}

	var tokenStr string

	// 优先从Authorization header获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenStr = authHeader[7:]
		}
	}

	// 其次从Cookie获取
	if tokenStr == "" {
		tokenStr, _ = c.Cookie("token")
	}

	// 最后从URL参数获取（WebSocket连接用）
	if tokenStr == "" {
		tokenStr = c.Query("token")
	}

	if tokenStr == "" {
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
	ctx := c.Request.Context()
	str, err := global.RDB.Get(ctx, fmt.Sprintf("user:token:%d", claims.UserID)).Result()
	if err != nil {
		response.FailWithError(c, enum.CodeUnauthorized, err)
		c.Abort()
		return
	}
	if str != tokenStr {
		response.FailWithCode(c, enum.CodeUnauthorized)
		c.Abort()
		return
	}
	// 存入上下文，后续路由可直接获取
	ctx = context.WithValue(ctx, "claims", claims)
	c.Request = c.Request.WithContext(ctx)
	c.Set("id", claims.UserID)
	c.Next()
}
