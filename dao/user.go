package dao

import (
	"LittleTalk/global"
	"LittleTalk/models"
	"context"
	"errors"
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

// GetByPhone 根据手机号查询用户
func GetByPhone(ctx context.Context, user *models.User, phone string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	err := global.DB.Where("phone = ?", phone).First(user).Error
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

// GetByEmail 根据邮箱查询用户
func GetByEmail(ctx context.Context, user *models.User, email string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	err := global.DB.Where("email = ?", email).First(user).Error
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}
