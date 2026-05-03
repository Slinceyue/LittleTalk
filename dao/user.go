package dao

import (
	"LittleTalk/global"
	"LittleTalk/models"
	"context"
	"time"
)

func CreatUser(ctx context.Context, user models.User) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	err := global.DB.Create(&user).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdateUser 更新用户信息
func UpdateUser(ctx context.Context, user models.User) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	err := global.DB.Save(&user).Error
	if err != nil {
		return err
	}
	return nil
}
