package tools_handler

import (
	"LittleTalk/api/request"
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/service"

	"github.com/gin-gonic/gin"
)

func (ToolsHandler) FileUploadHandler(c *gin.Context) {
	// 获取文件类型 (前端传 type，后端用 file_type)
	fileType := c.PostForm("type")
	if fileType == "" {
		fileType = c.PostForm("file_type")
	}
	if fileType == "" {
		response.FailWithCode(c, enum.CodeInvalidParam)
		return
	}

	fileUploadReq := request.FileUploadRequest{
		FileType: enum.FileType(fileType),
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithError(c, enum.CodeFileLoadFail, err)
		return
	}
	str, err := service.UploadFile(c.Request.Context(), fileUploadReq, file, c.GetUint("id"))
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
		return
	}
	c.File(path)
}
