package fileutils

import (
	"fmt"
	"os"
	"privacycheck/pkg/logging"
)

// FileInfo 表示文件信息
type FileInfo struct {
	Path     string
	Size     int64
	Encoding string
}

// PathToFileInfo 将文件路径转换为FileInfo结构体
// 内部整合路径存在性检查、是否为目录判断，避免重复的文件状态查询
func PathToFileInfo(path string) (FileInfo, error) {
	// 一次os.Stat调用获取所有文件基本信息（替代原PathExists的功能）
	fileStat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return FileInfo{}, fmt.Errorf("file %s not exists", path)
		}
		// 其他错误（如权限问题）
		return FileInfo{}, fmt.Errorf("obtain file %s info error: %w", path, err)
	}

	// 判断是否为目录（原PathExists的isDir判断）
	if fileStat.IsDir() {
		return FileInfo{}, fmt.Errorf("file %s is a directory", path)
	}

	// 检测文件编码（保持原逻辑）
	encoding, err := DetectFileEncoding(path)
	if err != nil {
		logging.Warnf("detect file %s encoding error:%v", path, err)
	}

	// 组装并返回FileInfo
	return FileInfo{
		Path:     path,
		Size:     fileStat.Size(),
		Encoding: encoding,
	}, nil
}
