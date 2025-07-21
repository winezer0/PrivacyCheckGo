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
func FilterByExtension(path string, excludeExtList []string) bool {
	// 若排除关键字列表为空，直接返回不排除
	if len(excludeExtList) == 0 {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	for _, excludeExt := range excludeExtList {
		if excludeExt == ext {
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

// FilterByPathKeys 检查路径是否包含任何排除关键字，若包含则返回true表示需要排除
func FilterByPathKeys(path string, excludeKeys []string) bool {
	// 若排除关键字列表为空，直接返回不排除
	if len(excludeKeys) == 0 {
		return false
	}
	// 遍历所有排除关键字，检查路径是否包含其中任何一个
	path = strings.ToLower(path)
	for _, key := range excludeKeys {
		if strings.Contains(path, key) {
			return true // 包含关键字，需要排除
		}
	}
	return false // 不包含任何关键字，无需排除
}

// GetFilesWithFilter 获取符合条件的文件列表
func GetFilesWithFilter(path string, excludeSuffixes, excludePathKeys []string, limitSize int) ([]string, error) {
	var files []string

	// 获取所有文件列表
	allFile, err := GetAllFilePaths(path)
	if err != nil && len(allFile) == 0 {
		return nil, err
	}

	// 过滤文件
	excludePathKeys = toLowerKeys(excludePathKeys)
	excludeSuffixes = toLowerKeys(excludeSuffixes)
	for _, file := range allFile {
		if FilterByExtension(file, excludeSuffixes) {
			continue
		}

		if FilterByPathKeys(file, excludePathKeys) {
			continue
		}

		if FileIsLarger(file, limitSize) {
			continue
		}

		files = append(files, file)
	}
	return files, err
}

// toLowerKeys 将排除关键字列表全部转为小写
func toLowerKeys(keys []string) []string {
	// 显式处理空列表，避免不必要的切片创建（虽然make空切片性能影响极小，但更直观）
	if len(keys) == 0 {
		return []string{}
	}

	lowerKeys := make([]string, len(keys))
	for i, key := range keys {
		lowerKeys[i] = strings.ToLower(key)
	}
	return lowerKeys
}
