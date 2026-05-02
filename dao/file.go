package dao

import (
	"LittleTalk/global"
	"LittleTalk/models"
	"context"
	"errors"
	"time"
)

func NewFile(ctx context.Context, file *models.File) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	err := global.DB.Create(&file).Error
	if err != nil {
		return errors.New("创建文件记录失败: " + err.Error())
	}
	return nil
}
