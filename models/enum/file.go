package enum

type FileType string

const (
	FileTypeImage FileType = "image" // 普通图片
	FileTypeFile  FileType = "file"  // 普通文件
)

func (f FileType) String() string {
	return string(f)
}

func IsValidUploadType(t string) bool {
	fileType := FileType(t)
	switch fileType {
	case FileTypeImage, FileTypeFile:
		return true
	default:
		return false
	}
}
