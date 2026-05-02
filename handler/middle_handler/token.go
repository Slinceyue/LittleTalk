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
