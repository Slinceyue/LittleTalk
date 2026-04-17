package dao

import (
	"LittleTalk/global"
	"LittleTalk/models"
)

func NewFile(file *models.File) error {
	err := global.DB.Create(file).Error
	return err
}
