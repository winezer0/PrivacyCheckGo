package fileutils

import (
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
	"io"
	"os"
	"strings"
)

// DetectFileEncoding 高效检测文件编码
// 优先检测BOM，然后使用chardet库进行内容分析
func DetectFileEncoding(filePath string) (string, error) {
	defaultEncode := "utf-8"

	file, err := os.Open(filePath)
	if err != nil {
		return defaultEncode, err
	}
	defer file.Close()

	// 读取足够的字节用于编码检测（4KB通常足够）
	buffer := make([]byte, 4096)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return defaultEncode, err
	}
	buffer = buffer[:n]

	// 先检测BOM（优先级最高）
	if encoding := detectBOM(buffer); encoding != "" {
		return encoding, nil
	}

	// 使用chardet库进行内容分析
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(buffer)
	if err != nil {
		return defaultEncode, err
	}

	// 只有当置信度足够高时才返回检测结果，否则返回默认值
	if result.Confidence > 50 {
		return result.Charset, nil
	}

	// 默认假设为UTF-8
	return defaultEncode, nil
}

// detectBOM 检测字节序列中的BOM标记
func detectBOM(data []byte) string {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return "utf-8"
	}
	if len(data) >= 2 {
		if data[0] == 0xFF && data[1] == 0xFE {
			if len(data) >= 4 && data[2] == 0x00 && data[3] == 0x00 {
				return "utf-32le"
			}
			return "utf-16le"
		}
		if data[0] == 0xFE && data[1] == 0xFF {
			if len(data) >= 4 && data[2] == 0x00 && data[3] == 0x00 {
				return "utf-32be"
			}
			return "utf-16be"
		}
		if len(data) >= 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0xFE && data[3] == 0xFF {
			return "utf-32be"
		}
		if len(data) >= 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0xFF && data[3] == 0xFE {
			return "utf-32le"
		}
	}
	return ""
}

// normalizedEncode 统一编码名称格式并返回对应的编码器
// 支持多种编码名称别名（如"utf8"等同于"utf-8"，"windows-936"等同于"gbk"）
// 若编码名称无法识别，默认返回UTF-8编码器
func normalizedEncode(encodeName string) encoding.Encoding {
	// 第一步：统一编码名称格式（规范化）
	normalizedName := strings.ToLower(encodeName)
	switch normalizedName {
	case "utf8", "utf-8":
		normalizedName = "utf-8"
	case "utf16", "utf-16":
		normalizedName = "utf-16"
	case "utf16le", "utf-16le":
		normalizedName = "utf-16le"
	case "utf16be", "utf-16be":
		normalizedName = "utf-16be"
	case "utf32", "utf-32":
		normalizedName = "utf-32"
	case "utf32le", "utf-32le":
		normalizedName = "utf-32le"
	case "utf32be", "utf-32be":
		normalizedName = "utf-32be"
	case "windows-936", "gbk": // windows-936是GBK的Windows编码名
		normalizedName = "gbk"
	case "gbk2312", "gb2312": // 处理常见的GB2312别名
		normalizedName = "gb2312"
	case "big5-hkscs", "big5": // big5-hkscs是Big5的扩展
		normalizedName = "big5"
		// 其他未匹配的编码名称保持原样，后续判断
	}

	// 第二步：根据规范化后的名称映射到对应的编码器
	switch normalizedName {
	case "utf-16":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	case "utf-16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	case "utf-16be":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	case "utf-32":
		return utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM)
	case "utf-32le":
		return utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM)
	case "utf-32be":
		return utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM)
	case "gbk":
		return simplifiedchinese.GBK
	case "gb2312":
		return simplifiedchinese.HZGB2312
	case "gb18030":
		return simplifiedchinese.GB18030
	case "big5":
		return traditionalchinese.Big5
	// 未识别的编码默认返回UTF-8编码器
	default:
		return unicode.UTF8
	}
}
