package response

import (
	"LittleTalk/models/enum"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (res Response) OKJson(c *gin.Context) {
	c.JSON(http.StatusOK, res)
}

func OK(c *gin.Context) {
	Response{Code: enum.CodeSuccess.Int(), Message: enum.CodeSuccess.Message(), Data: nil}.OKJson(c)
}

func OKWithData(c *gin.Context, data interface{}) {
	Response{Code: enum.CodeSuccess.Int(), Message: enum.CodeSuccess.Message(), Data: data}.OKJson(c)
}

func OKWithMsg(c *gin.Context, msg string) {
	Response{Code: enum.CodeSuccess.Int(), Message: msg}.OKJson(c)
}
