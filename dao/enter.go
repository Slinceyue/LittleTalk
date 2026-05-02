package dao

import (
	"LittleTalk/global"
	"context"
	"time"
)

func Creat[T any](ctx context.Context, model T) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	return global.DB.Create(&model).Error
}
func Get[T any](ctx context.Context, where T) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	var model T
	err := global.DB.Where(where).First(&model).Error
	return model, err
}
func GetByKey[T any](ctx context.Context, where T, keyType string, key string) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	var model T
	err := global.DB.Where(where).Where(keyType+" LIKE ?", "%"+key+"%").First(&model).Error
	return model, err
}
func GetByID[T any](ctx context.Context, model T, id uint) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	var _model T
	err := global.DB.Model(&model).First(&_model, id).Error
	return _model, err
}
func UpDateByID[T any](ctx context.Context, model T, id uint) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	err := global.DB.Model(&model).Where("id = ?", id).Updates(model).Error
	return err
}
func DeleteByID[T any](ctx context.Context, model T, id uint) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	err := global.DB.Delete(&model, id).Error
	return err
}
