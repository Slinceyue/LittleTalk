package request

import "LittleTalk/models/enum"

type FileUploadRequest struct {
	FileType enum.FileType `form:"file_type" json:"file_type"binding:"required" label:"文件类型"`
}
type FileReadRequest struct {
	FromID   uint          `form:"from_id" json:"from_id"binding:"required" label:"上传者ID"`
	FileName string        `form:"file_name" json:"file_name"binding:"required" label:"文件名"`
	FileType enum.FileType `form:"file_type" json:"file_type"binding:"required" label:"文件类型"`
}
