package fileutils

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultExcludeExt 默认排除的文件扩展名
var DefaultExcludeExt = []string{
	".tmp", ".exe", ".bin", ".dll", ".elf", ".so", ".dylib",
	".zip", ".rar", ".7z", ".gz", ".bz2", ".tar", ".xz",
	".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".tif",
	".woff", ".woff2", ".ttf", ".otf", ".eot",
	".mp3", ".mp4", ".avi", ".mov", ".wmv", ".flv",
	".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
}

// FilterByExtension 检查文件是否应该根据扩展名被排除
// 返回true表示需要排除，false表示保留
func FilterByExtension(path string, excludeExt []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range excludeExt {
		if strings.ToLower(e) == ext {
			return true
		}
	}
	return false
}

// FileIsLarger 检查文件是否超过指定大小（MB）
func FileIsLarger(filePath string, limitSize int) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.Size() > int64(limitSize*1024*1024)
}
