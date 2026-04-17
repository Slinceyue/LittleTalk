package tools_handler

import (
	"LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/service"

	"github.com/gin-gonic/gin"
)

func (ToolsHandler) FileUploadHandler(c *gin.Context) {
	var filereq request.FileUploadRequest
	if err := c.ShouldBind(&filereq); err != nil {
		response.FailWithCode(c, enum.CodeInvalidParam)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithError(c, enum.CodeFileLoadFail, err)
		return
	}
	str, err := service.UploadFile(filereq, file, c.GetUint("id"))
	if err != nil {
		response.FailWithError(c, enum.CodeFileUploadFail, err)
		return
	}
	response.OKWithData(c, str)
	return
}
