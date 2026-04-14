package models

import "LittleTalk/models/enum"

type File struct {
	Model
	FileName string        `gorm:"size:255;not null" json:"file_name"`
	FilePath string        `gorm:"size:255;not null" json:"file_path"`
	FileType enum.FileType `gorm:"default:file" json:"file_type"`
}
