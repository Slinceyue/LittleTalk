package dao

import (
	"errors"
	"LittleTalk/global"
	"context"
	"strings"
	"time"
)

// 预定义错误
var (
	// ErrRecordNotFound 记录不存在
	ErrRecordNotFound = errors.New("record not found")
	// ErrDuplicateEntry 重复条目（如用户名已存在）
	ErrDuplicateEntry = errors.New("duplicate entry")
)

// IsDuplicateEntry 判断错误是否为重复条目错误
func IsDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate entry")
}

// Creat 创建记录
func Creat[T any](ctx context.Context, model T) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	return global.DB.Create(&model).Error
}

// Get 根据条件查询单条记录
func Get[T any](ctx context.Context, where T) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	var model T
	err := global.DB.Where(where).First(&model).Error
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return model, ErrRecordNotFound
		}
		return model, err
	}
	return model, nil
}

// GetByKey 根据键值模糊查询
func GetByKey[T any](ctx context.Context, where T, keyType string, key string) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	var model T
	err := global.DB.Where(where).Where(keyType+" LIKE ?", "%"+key+"%").First(&model).Error
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return model, ErrRecordNotFound
		}
		return model, err
	}
	return model, nil
}

// GetByID 根据ID查询单条记录
func GetByID[T any](ctx context.Context, model T, id uint) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	var _model T
	err := global.DB.Model(&model).First(&_model, id).Error
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return _model, ErrRecordNotFound
		}
		return _model, err
	}
	return _model, nil
}

// GetByIDs 批量根据ID查询
func GetByIDs[T any](ctx context.Context, model *[]T, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	err := global.DB.Where("id IN ?", ids).Find(model).Error
	return err
}

// UpDateByID 根据ID更新记录
func UpDateByID[T any](ctx context.Context, model T, id uint) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	err := global.DB.Model(&model).Where("id = ?", id).Updates(model).Error
	return err
}

// DeleteByID 根据ID删除记录
func DeleteByID[T any](ctx context.Context, model T, id uint) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	err := global.DB.Delete(&model, id).Error
	return err
}
