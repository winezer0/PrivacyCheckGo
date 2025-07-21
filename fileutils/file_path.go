package fileutils

import (
	"os"
	"path/filepath"
)

// PathExists 检查路径是否存在，并返回是否为目录
func PathExists(path string) (exists bool, isDir bool, err error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, info.IsDir(), nil
}

// FileExists 检查文件是否存在
func FileExists(path string) bool {
	exists, isDir, _ := PathExists(path)
	return exists && !isDir
}

// DirExists 检查目录是否存在
func DirExists(path string) bool {
	exists, isDir, _ := PathExists(path)
	return exists && isDir
}

// GetAllFilePaths 获取指定路径下的所有文件路径
// 如果输入是文件路径，则直接返回包含该文件路径的切片
// 如果输入是目录路径，则返回该目录下所有文件（包括子目录中的文件）的路径
func GetAllFilePaths(path string) ([]string, error) {
	// 获取路径信息
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// 如果是文件，直接返回包含该文件路径的切片
	if !info.IsDir() {
		return []string{path}, nil
	}

	// 如果是目录，遍历所有文件
	var filePaths []string
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // 传递错误
		}

		// 只添加文件，跳过目录
		if !info.IsDir() {
			filePaths = append(filePaths, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return filePaths, nil
}
