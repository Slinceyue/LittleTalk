package service

import (
	"LittleTalk/api/response"
	"LittleTalk/models/enum"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func UploadGeneric(c *gin.Context) {
	uploadType := c.PostForm("type")
	if !enum.IsValidUploadType(uploadType) {
		response.FailWithCode(c, enum.CodeFileTypeWrong)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithCode(c, enum.CodeFileWrong)
	}
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		c.JSON(400, gin.H{"error": "文件无后缀"})
		return
	}
	var maxSize int64
	var allowExts map[string]bool

	switch uploadType {
	case "avatar":
		// 头像：小尺寸 + 图片
		maxSize = 2 << 20 // 2MB
		allowExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true}

	case "image":
		// 普通图片：稍大
		maxSize = 10 << 20 // 10MB
		allowExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}

		switch uploadType {
}
