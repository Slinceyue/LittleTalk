package tools_handler

import (
	"LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/service"

	"github.com/gin-gonic/gin"
)

func (ToolsHandler) FileUploadHandler(c *gin.Context) {
	var fileUploadReq request.FileUploadRequest
	if err := c.ShouldBind(&fileUploadReq); err != nil {
		response.FailWithCode(c, enum.CodeInvalidParam)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithError(c, enum.CodeFileLoadFail, err)
		return
	}
	str, err := service.UploadFile(fileUploadReq, file, c.GetUint("id"))
	if err != nil {
		response.FailWithError(c, enum.CodeFileUploadFail, err)
		return
	}
	response.OKWithData(c, str)
	return
}

func (ToolsHandler) ReadFileHandler(c *gin.Context) {
	var fileUploadReq request.FileReadRequest
	if err := c.ShouldBind(&fileUploadReq); err != nil {
		response.FailWithCode(c, enum.CodeInvalidParam)
		return
	}
	var path, err = service.UploadPath(fileUploadReq)
	if err != nil {
		response.FailWithError(c, enum.CodeFileLoadFail, err)
	}
	c.File(path)
}
