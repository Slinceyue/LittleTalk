package user_handler

import (
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/service"

	"github.com/gin-gonic/gin"
)

func (UserHandler) AvatarUpload(c *gin.Context) {
	file, err := c.FormFile("avatar")
	if err != nil {
		response.FailWithMsg(c, enum.CodeInvalidParam, "请选择头像文件")
		return
	}
	err = service.UploadAvatar(c.Request.Context(), file, c.GetUint("id"))
	if err != nil {
		response.FailWithError(c, enum.CodeFileUploadFail, err)
		return
	}
	response.OKWithMsg(c, "头像上传成功")
}
