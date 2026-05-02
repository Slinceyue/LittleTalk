package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

func UploadAvatar(ctx context.Context, file *multipart.FileHeader, id uint) error {
	// 1. 文件判空
	if file == nil || file.Size == 0 {
		return errors.New("文件不能为空1")
	}

	// 2. 文件大小限制（40MB）
	if file.Size > 40<<20 {
		return errors.New("文件不能超过40MB")
	}

	// 3. 获取后缀（统一小写，带点），白名单校验（修复点1）
	ext := strings.ToLower(filepath.Ext(file.Filename))
	// 正确判断：后缀是 .jpg（带点）
	if ext != ".jpg" {
		return errors.New("文件格式错误，仅支持jpg")
	}

	// 4. 目录创建（递归创建，权限0755）
	avatarDir := "static/avatar"
	err := os.MkdirAll(avatarDir, 0755)
	if err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 5. 安全拼接路径（跨平台，修复点3）
	fileName := fmt.Sprintf("%d%s", id, ext)
	savePath := filepath.Join(avatarDir, fileName)

	// 6. 删除旧头像：只忽略「文件不存在」，其他错误返回（修复点2）
	err = os.Remove(savePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除旧头像失败: %w", err)
	}

	// 7. 打开上传文件流
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	// 8. 创建新文件（覆盖写，权限0644）
	dst, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	// 9. 拷贝文件内容
	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("文件上传失败: %w", err)
	}

	return nil
}
