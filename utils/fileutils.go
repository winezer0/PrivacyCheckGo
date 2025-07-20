package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"privacycheck/core"
	"privacycheck/logging"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// 默认排除的文件扩展名
var DefaultExcludeExt = []string{
	".tmp", ".exe", ".bin", ".dll", ".elf", ".so", ".dylib",
	".zip", ".rar", ".7z", ".gz", ".bz2", ".tar", ".xz",
	".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".tif",
	".woff", ".woff2", ".ttf", ".otf", ".eot",
	".mp3", ".mp4", ".avi", ".mov", ".wmv", ".flv",
	".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
}

// GetFilesWithFilter 获取符合条件的文件列表
func GetFilesWithFilter(targetPath string, excludeExt []string, limitSizeMB int) ([]core.FileInfo, error) {
	var files []core.FileInfo
	
	// 合并默认排除扩展名和用户指定的扩展名
	allExcludeExt := append(DefaultExcludeExt, excludeExt...)
	excludeMap := make(map[string]bool)
	for _, ext := range allExcludeExt {
		excludeMap[strings.ToLower(ext)] = true
	}

	limitSizeBytes := int64(limitSizeMB) * 1024 * 1024

	// 检查是否为单个文件
	if info, err := os.Stat(targetPath); err == nil && !info.IsDir() {
		if !shouldExcludeFile(targetPath, info, excludeMap, limitSizeBytes) {
			encoding := DetectFileEncoding(targetPath)
			files = append(files, core.FileInfo{
				Path:     targetPath,
				Size:     info.Size(),
				Encoding: encoding,
			})
		}
		return files, nil
	}

	// 遍历目录
	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logging.Warnf("访问文件失败 %s: %v", path, err)
			return nil // 继续处理其他文件
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查是否应该排除此文件
		if shouldExcludeFile(path, info, excludeMap, limitSizeBytes) {
			return nil
		}

		// 检测文件编码
		encoding := DetectFileEncoding(path)
		
		files = append(files, core.FileInfo{
			Path:     path,
			Size:     info.Size(),
			Encoding: encoding,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %w", err)
	}

	return files, nil
}

// shouldExcludeFile 检查是否应该排除文件
func shouldExcludeFile(path string, info os.FileInfo, excludeMap map[string]bool, limitSizeBytes int64) bool {
	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(path))
	if excludeMap[ext] {
		return true
	}

	// 检查文件大小
	if limitSizeBytes > 0 && info.Size() > limitSizeBytes {
		return true
	}

	return false
}

// DetectFileEncoding 检测文件编码
func DetectFileEncoding(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return "utf-8"
	}
	defer file.Close()

	// 读取文件前1024字节用于编码检测
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "utf-8"
	}

	buffer = buffer[:n]
	return detectEncoding(buffer)
}

// detectEncoding 根据字节内容检测编码
func detectEncoding(data []byte) string {
	// 检查UTF-8 BOM
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return "utf-8-sig"
	}

	// 检查UTF-16 BOM
	if len(data) >= 2 {
		if data[0] == 0xFF && data[1] == 0xFE {
			return "utf-16le"
		}
		if data[0] == 0xFE && data[1] == 0xFF {
			return "utf-16be"
		}
	}

	// 尝试不同编码
	encodings := []string{"utf-8", "gbk", "gb2312", "gb18030", "big5"}
	
	for _, encoding := range encodings {
		if isValidEncoding(data, encoding) {
			return encoding
		}
	}

	// 默认返回UTF-8
	return "utf-8"
}

// isValidEncoding 检查数据是否符合指定编码
func isValidEncoding(data []byte, encoding string) bool {
	var decoder transform.Transformer
	
	switch strings.ToLower(encoding) {
	case "utf-8":
		// UTF-8验证
		return isValidUTF8(data)
	case "gbk":
		decoder = simplifiedchinese.GBK.NewDecoder()
	case "gb2312":
		decoder = simplifiedchinese.HZGB2312.NewDecoder()
	case "gb18030":
		decoder = simplifiedchinese.GB18030.NewDecoder()
	case "big5":
		decoder = traditionalchinese.Big5.NewDecoder()
	default:
		return false
	}

	if decoder != nil {
		_, _, err := transform.Bytes(decoder, data)
		return err == nil
	}

	return false
}

// isValidUTF8 检查是否为有效的UTF-8编码
func isValidUTF8(data []byte) bool {
	for len(data) > 0 {
		r, size := decodeUTF8Rune(data)
		if r == 0xFFFD && size == 1 {
			return false
		}
		data = data[size:]
	}
	return true
}

// decodeUTF8Rune 解码UTF-8字符
func decodeUTF8Rune(data []byte) (rune, int) {
	if len(data) == 0 {
		return 0, 0
	}

	b := data[0]
	if b < 0x80 {
		return rune(b), 1
	}

	if b < 0xC0 {
		return 0xFFFD, 1
	}

	if b < 0xE0 {
		if len(data) < 2 {
			return 0xFFFD, 1
		}
		return rune(b&0x1F)<<6 | rune(data[1]&0x3F), 2
	}

	if b < 0xF0 {
		if len(data) < 3 {
			return 0xFFFD, 1
		}
		return rune(b&0x0F)<<12 | rune(data[1]&0x3F)<<6 | rune(data[2]&0x3F), 3
	}

	if b < 0xF8 {
		if len(data) < 4 {
			return 0xFFFD, 1
		}
		return rune(b&0x07)<<18 | rune(data[1]&0x3F)<<12 | rune(data[2]&0x3F)<<6 | rune(data[3]&0x3F), 4
	}

	return 0xFFFD, 1
}

// ReadFileSafe 安全读取文件内容
func ReadFileSafe(filePath, encoding string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file

	// 根据编码设置解码器
	switch strings.ToLower(encoding) {
	case "utf-8-sig":
		// 跳过UTF-8 BOM
		bom := make([]byte, 3)
		file.Read(bom)
		if !(bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF) {
			file.Seek(0, 0) // 如果不是BOM，重置文件指针
		}
		reader = file
	case "gbk":
		reader = transform.NewReader(file, simplifiedchinese.GBK.NewDecoder())
	case "gb2312":
		reader = transform.NewReader(file, simplifiedchinese.HZGB2312.NewDecoder())
	case "gb18030":
		reader = transform.NewReader(file, simplifiedchinese.GB18030.NewDecoder())
	case "big5":
		reader = transform.NewReader(file, traditionalchinese.Big5.NewDecoder())
	case "utf-16le":
		reader = transform.NewReader(file, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder())
	case "utf-16be":
		reader = transform.NewReader(file, unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder())
	case "latin1", "iso-8859-1":
		reader = transform.NewReader(file, charmap.ISO8859_1.NewDecoder())
	}

	// 读取内容
	content, err := io.ReadAll(reader)
	if err != nil {
		// 如果解码失败，尝试强制使用UTF-8并忽略错误
		file.Seek(0, 0)
		content, err = io.ReadAll(file)
		if err != nil {
			return "", fmt.Errorf("读取文件内容失败: %w", err)
		}
		// 替换无效字符
		return strings.ToValidUTF8(string(content), "�"), nil
	}

	return string(content), nil
}

// ReadFileInChunks 分块读取文件
func ReadFileInChunks(filePath, encoding string, chunkSize int, callback func(chunk string) error) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file

	// 设置解码器（简化版本，仅支持常用编码）
	switch strings.ToLower(encoding) {
	case "gbk":
		reader = transform.NewReader(file, simplifiedchinese.GBK.NewDecoder())
	case "gb18030":
		reader = transform.NewReader(file, simplifiedchinese.GB18030.NewDecoder())
	}

	scanner := bufio.NewScanner(reader)
	buffer := make([]byte, chunkSize)
	scanner.Buffer(buffer, chunkSize)

	var chunk strings.Builder
	currentSize := 0

	for scanner.Scan() {
		line := scanner.Text() + "\n"
		chunk.WriteString(line)
		currentSize += len(line)

		if currentSize >= chunkSize {
			if err := callback(chunk.String()); err != nil {
				return err
			}
			chunk.Reset()
			currentSize = 0
		}
	}

	// 处理最后一个块
	if chunk.Len() > 0 {
		if err := callback(chunk.String()); err != nil {
			return err
		}
	}

	return scanner.Err()
}
