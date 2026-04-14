package response

import (
	"LittleTalk/models/enum"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (res Response) FailJson(c *gin.Context) {
	c.JSON(http.StatusBadRequest, res)
}

func FailWithCode(c *gin.Context, code enum.ResCode) {
	Response{Code: code.Int(), Message: code.Message()}.FailJson(c)
}

func FailWithError(c *gin.Context, code enum.ResCode, e error) {
	Response{Code: code.Int(), Message: e.Error()}.FailJson(c)
}

func FailWithMsg(c *gin.Context, code enum.ResCode, msg string) {
	Response{Code: code.Int(), Message: msg}.FailJson(c)
}
