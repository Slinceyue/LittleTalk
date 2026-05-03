package user_handler

import (
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"LittleTalk/service"

	"github.com/gin-gonic/gin"
)

func (UserHandler) AvatarUpload(c *gin.Context) {
	file, _ := c.FormFile("avatar")
	err := service.UploadAvatar(c.Request.Context(), file, c.GetUint("id"))
	if err != nil {
		response.FailWithError(c, enum.CodeFileLoadFail, err)
		return
	}
	response.OKWithMsg(c, "头像上传成功")
}
