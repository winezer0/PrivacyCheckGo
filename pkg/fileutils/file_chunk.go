package fileutils

import (
	"bufio"
	"golang.org/x/text/transform"
	"io"
	"os"
	"strings"
)

// ChunkInfo 表示文件块信息
type ChunkInfo struct {
	Content     string // 块内容
	StartLine   int    // 块起始行号
	EndLine     int    // 块结束行号
	StartOffset int64  // 块在文件中的起始字节偏移
	EndOffset   int64  // 块在文件中的结束字节偏移
}

// ReadFileByChunk 按指定大小读取文件块，保留行号功能
// chunkSize: 每个块的大小（字节）
// handler: 处理每个块的回调函数
func ReadFileByChunk(filePath, encode string, chunkSize int, handler func(chunk ChunkInfo) error) error {
	// 如果未指定编码，则自动检测
	if encode == "" {
		detectedEnc, err := DetectFileEncoding(filePath)
		if err == nil && detectedEnc != "" {
			encode = detectedEnc
		} else {
			encode = "utf-8"
		}
	}

	// 获取编码器
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

	// 创建缓冲读取器
	bufReader := bufio.NewReader(reader)

	var currentLine = 1
	var currentOffset int64 = 0
	var buffer []byte
	var remainingData []byte

	for {
		// 读取指定大小的数据块
		chunk := make([]byte, chunkSize)
		n, err := bufReader.Read(chunk)
		if n == 0 {
			break
		}

		// 将剩余数据和新读取的数据合并
		if len(remainingData) > 0 {
			buffer = append(remainingData, chunk[:n]...)
		} else {
			buffer = chunk[:n]
		}

		// 找到最后一个完整行的位置
		content := string(buffer)
		lastNewlineIndex := strings.LastIndex(content, "\n")

		var chunkContent string
		var nextRemainingData []byte

		if lastNewlineIndex == -1 {
			// 没有找到换行符，整个块都是一行的一部分
			if err == io.EOF {
				// 如果是文件末尾，包含所有内容
				chunkContent = content
				nextRemainingData = nil
			} else {
				// 不是文件末尾，保留所有数据到下一次
				remainingData = buffer
				continue
			}
		} else {
			// 找到换行符，分割内容
			chunkContent = content[:lastNewlineIndex+1]
			if lastNewlineIndex+1 < len(content) {
				nextRemainingData = []byte(content[lastNewlineIndex+1:])
			} else {
				nextRemainingData = nil
			}
		}

		// 计算行号
		startLine := currentLine
		lineCount := strings.Count(chunkContent, "\n")
		endLine := currentLine + lineCount
		if len(chunkContent) > 0 && !strings.HasSuffix(chunkContent, "\n") {
			endLine-- // 如果最后没有换行符，行号不增加
		}

		// 创建块信息
		chunkInfo := ChunkInfo{
			Content:     chunkContent,
			StartLine:   startLine,
			EndLine:     endLine,
			StartOffset: currentOffset,
			EndOffset:   currentOffset + int64(len(chunkContent)),
		}

		// 调用处理函数
		if err := handler(chunkInfo); err != nil {
			return err
		}

		// 更新状态
		currentLine = endLine
		if len(chunkContent) > 0 && strings.HasSuffix(chunkContent, "\n") {
			currentLine++
		}
		currentOffset += int64(len(chunkContent))
		remainingData = nextRemainingData

		// 如果到达文件末尾，退出循环
		if err == io.EOF {
			break
		}
	}

	// 处理剩余数据
	if len(remainingData) > 0 {
		chunkInfo := ChunkInfo{
			Content:     string(remainingData),
			StartLine:   currentLine,
			EndLine:     currentLine,
			StartOffset: currentOffset,
			EndOffset:   currentOffset + int64(len(remainingData)),
		}
		return handler(chunkInfo)
	}

	return nil
}
