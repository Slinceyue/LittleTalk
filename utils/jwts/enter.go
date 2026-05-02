package jwts

import (
	"LittleTalk/global"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"userID"`
	UserName string `json:"username"`
	Role     int8   `json:"role"`
}
type MyClaims struct {
	Claims
	jwt.RegisteredClaims
}

func GetToken(claims Claims) (string, error) {
	cla := MyClaims{
		Claims: claims,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(global.Config.Jwt.Expire) * 24 * time.Hour)),
			Issuer:    global.Config.Jwt.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cla)
	return token.SignedString([]byte(global.Config.Jwt.Secret))
}

func ParseToken(TokenString string) (*MyClaims, error) {
	if TokenString == "" {
		return nil, errors.New("请登录")
	}
	token, err := jwt.ParseWithClaims(TokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(global.Config.Jwt.Secret), nil
	})
	if err != nil {
		// 优先判断标准错误（v5 推荐）
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, errors.New("token已过期")
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, errors.New("token格式错误")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, errors.New("token签名无效")
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, errors.New("token尚未生效")
		case errors.Is(err, jwt.ErrTokenInvalidIssuer):
			return nil, errors.New("token签发者无效")
		default:
			return nil, errors.New("token无效")
		}
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

func ParseTokenByGen(c *gin.Context) (*MyClaims, error) {
	token := c.GetHeader("token")
	if token == "" {
		token = c.Query("token")
	}

	return ParseToken(token)
}
