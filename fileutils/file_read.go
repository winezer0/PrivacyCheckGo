package fileutils

import (
	"bufio"
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
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// 直接读取UTF-8文件为字符串
func readFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// ReadLargeFileWithEncoding 读取大文件，适合处理几MB以上的文件 会逐行处理以减少内存占用
func ReadLargeFileWithEncoding(filePath, encode string, handler func(line string) error) error {
	// 如果未指定编码，则自动检测
	if encode == "" {
		detectedEnc, err := DetectFileEncoding(filePath)
		if err != nil {
			return err
		}
		encode = detectedEnc
	}

	// 获取编码器，复用getEncoding函数
	enc := normalizedEncode(encode)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var reader io.Reader = file

	// 如果不是UTF-8编码，创建解码读取器
	if encode != "utf-8" {
		decoder := enc.NewDecoder()
		reader = transform.NewReader(file, ignoreErrors(decoder))
	}

	// 创建扫描器并逐行处理
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := handler(scanner.Text()); err != nil {
			return err
		}
	}

	return scanner.Err()
}
