package fileutils

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
)

// WriteJSONFile 写入JSON文件
func WriteJSONFile(filename string, data interface{}) error {
	file, err := CreateFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// WriteCSVFile 写入CSV文件
func WriteCSVFile(filename string, headers []string, records [][]string) error {
	file, err := CreateFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return err
		}
	}

	// 写入数据
	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// WriteFile 写入文件内容，如果目录不存在会自动创建
func WriteFile(filePath string, data []byte) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := EnsureDir(dir); err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(filePath, data, 0644)
}

// CreateFile 创建文件，如果目录不存在会自动创建
func CreateFile(filePath string) (*os.File, error) {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// 创建文件
	return os.Create(filePath)
}
