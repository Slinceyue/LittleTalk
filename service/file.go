package service

import (
	"LittleTalk/api/request"
	"LittleTalk/dao"
	"LittleTalk/models"
	"LittleTalk/models/enum"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

func uploadPath(fileType enum.FileType, id uint) (string, error) {
	var path string
	if !enum.IsValidUploadType(fileType.String()) {
		return "", errors.New(enum.CodeFileTypeWrong.String())
	}
	path = fmt.Sprintf("static/uploads/%s/%d", fileType.String(), id)
	return path, nil
}

func UploadFile(req request.FileUploadRequest, file *multipart.FileHeader, id uint) (string, error) {
	// 1. 文件判空
	if file == nil || file.Size == 0 {
		return "", errors.New("文件不能为空")
	}
	if file.Size > 40<<20 {
		return "", errors.New("文件过大")
	}
	// 2. 获取存储目录
	filePath, err := uploadPath(req.FileType, id)
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(filePath, 0755)
	if err != nil {
		return "", errors.New("创建目录失败")
	}
	// 4. 获取文件后缀
	ext := filepath.Ext(file.Filename)

	// 5. 生成唯一文件名（时间戳+微秒，永不重复）
	uniqueName := fmt.Sprintf("%d%s", time.Now().UnixMicro(), ext)

	saveFullPath := fmt.Sprintf("%s/%s", filePath, uniqueName)

	src, err := file.Open()
	if err != nil {
		return "", errors.New("打开文件失败")
	}
	defer src.Close()
	// 8. 创建目标文件
	dst, err := os.Create(saveFullPath)
	if err != nil {
		return "", errors.New("创建文件失败")
	}
	defer dst.Close()

	// 9. 写入文件
	_, err = io.Copy(dst, src)
	if err != nil {
		return "", errors.New("文件上传失败")
	}
	model := models.File{
		UserID:    id,
		FileType:  req.FileType.String(),
		RealName:  file.Filename,
		SavedName: uniqueName,
		Src:       saveFullPath,
		Size:      file.Size,
	}
	err = dao.NewFile(&model)
	if err != nil {
		return "", errors.New(enum.CodeFileUploadFail.String())
	}
	// 10. 返回可访问的路径（给前端/数据库存）
	return saveFullPath, nil
}
