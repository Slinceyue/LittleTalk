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

	"LittleTalk/cache"
	"LittleTalk/global"
	"LittleTalk/models"
)

func UploadAvatar(ctx context.Context, file *multipart.FileHeader, id uint) error {
	// 1. 文件判空
	if file == nil || file.Size == 0 {
		return errors.New("文件不能为空")
	}

	// 2. 文件大小限制（2MB）
	if file.Size > 2<<20 {
		return errors.New("文件不能超过2MB")
	}

	// 3. 获取后缀，白名单校验（支持多种图片格式）
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		return errors.New("仅支持 jpg、png、gif、webp 格式")
	}

	// 4. 目录创建（递归创建，权限0755）
	avatarDir := "static/avatar"
	err := os.MkdirAll(avatarDir, 0755)
	if err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 5. 安全拼接路径（跨平台）
	fileName := fmt.Sprintf("%d%s", id, ext)
	savePath := filepath.Join(avatarDir, fileName)

	// 6. 删除旧头像：只忽略「文件不存在」，其他错误返回
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

	// 10. 更新数据库中的用户头像路径
	avatarURL := "/static/avatar/" + fileName
	err = global.DB.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("avatar", avatarURL).Error
	if err != nil {
		return fmt.Errorf("更新头像路径失败: %w", err)
	}

	// 11. 清除用户信息缓存，确保下次获取到最新头像
	_ = cache.DelUserInfoCache(ctx, id)

	return nil
}
