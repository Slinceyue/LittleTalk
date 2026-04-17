package request

import "LittleTalk/models/enum"

type FileUploadRequest struct {
	FileType enum.FileType `form:"file_type" json:"file_type"binding:"required" label:"文件类型"`
}
