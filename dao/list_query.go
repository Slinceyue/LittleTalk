package dao

import (
	"LittleTalk/global"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type PageInfo struct {
	Limit int    `form:"limit"`
	Page  int    `form:"page"`
	Key   string `form:"key"`
	Order string `form:"order"`
}

func (p PageInfo) GetPage() int {
	if p.Page >= 20 || p.Page <= 0 {
		return 1
	}
	return p.Page
}

func (p PageInfo) GetLimit() int {
	if p.Limit <= 0 || p.Limit > 100 {
		return 10
	}
	return p.Limit
}

func (p PageInfo) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

type Options struct {
	PageInfo     PageInfo
	Likes        []string
	Preloads     []string
	Where        *gorm.DB
	Debug        bool
	DefaultOrder string
}

func ListQuery[T any](ctx context.Context, model T, option Options) (list []T, count int, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(global.Config.System.ServerTimeout)*time.Second)
	defer cancel()
	//基础查询
	query := global.DB.Model(model).Where(model)

	if option.Debug {
		query = query.Debug()
	}
	//模糊匹配

	if len(option.Likes) > 0 && option.PageInfo.Key != "" {
		likes := global.DB.Where("")
		for _, column := range option.Likes {
			likes.Or(fmt.Sprintf("%s LIKE ?", column), fmt.Sprintf("%%%s%%", option.PageInfo.Key))
		}
		query = query.Where(likes)
	}
	//定制化查询
	if option.Where != nil {
		query = query.Where(option.Where)
	}

	for _, preload := range option.Preloads {
		query = query.Preload(preload)
	}

	//查总数
	var _c int64
	query.Count(&_c)
	count = int(_c)
	//分页
	limit := option.PageInfo.GetLimit()
	offset := option.PageInfo.GetOffset()
	//排序
	if option.PageInfo.Order != "" {
		query = query.Order(option.PageInfo.Order)
	} else {
		query = query.Order(option.DefaultOrder)
	}
	err = query.Offset(offset).Limit(limit).Find(&list).Error
	return
}
