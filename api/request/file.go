package request

import "LittleTalk/models/enum"

type FileRequest struct {
	FileType enum.FileType `json:"file_type"binding:"required" label:"文件类型"`
}
