package dao

import (
	"LittleTalk/global"
	"LittleTalk/models"
	"errors"
)

func NewFile(file *models.File) error {
	err := global.DB.Create(&file).Error
	if err != nil {
		return errors.New("创建文件记录失败: " + err.Error())
	}
	return nil
}
