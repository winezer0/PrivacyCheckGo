package fileutils

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
	"io"
	"os"
)

// ReadFileWithEncoding 读取指定编码的文件内容并转换为UTF-8字符串
// 如果encoding为空或不匹配任何已知编码，则默认以UTF-8读取
// 读取时会忽略解码错误，类似Python的errors=ignore
type ignoreErrorTransformer struct {
	transform.Transformer
}

// Transform 实现Transform方法，忽略解码错误
func (t ignoreErrorTransformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	nDst, nSrc, err = t.Transformer.Transform(dst, src, atEOF)
	if err != nil {
		// 忽略错误，继续处理
		return nDst, nSrc, nil
	}
	return nDst, nSrc, err
}

// 辅助函数：创建忽略错误的转换器
func ignoreErrors(t transform.Transformer) transform.Transformer {
	return ignoreErrorTransformer{t}
}

// ReadFileWithEncoding 读取指定编码的文件内容并转换为UTF-8字符串
// 如果encoding为空或不匹配任何已知编码，则默认以UTF-8读取
// 读取时会忽略解码错误，类似Python的errors=ignore
func ReadFileWithEncoding(filePath, encode string) (string, error) {
	// 如果未指定编码，则自动检测
	if encode == "" {
		detectedEnc, err := DetectFileEncoding(filePath)
		if err == nil && detectedEnc != "" {
			encode = detectedEnc
		} else {
			// 检测失败，默认使用UTF-8
			encode = "utf-8"
		}
	}
	// 获取编码器
	return readFileWithEncoding(filePath, normalizedEncode(encode))
}

// readFileWithEncoding 使用指定的编码器读取文件并解码为UTF-8字符串
// 忽略解码错误，类似Python的errors=ignore
func readFileWithEncoding(filePath string, enc encoding.Encoding) (string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 创建解码读取器，并使用自定义的错误忽略转换器
	decoder := enc.NewDecoder()
	reader := transform.NewReader(file, ignoreErrors(decoder))

	// 读取并解码内容
	content, err := io.ReadAll(reader)
	if err != nil && len(content) == 0 {
		return "", err
	}

	// 强制将字节切片转换为字符串，即使包含无效 UTF-8
	return string(content), nil
}

// ReadFile 读取文件内容
func ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
