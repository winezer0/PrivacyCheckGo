package fileprocessor

import (
	"fmt"
	"os"
	"path/filepath"
	"privacycheck/pkg/fileutils"
	"privacycheck/pkg/logging"
	"strings"
	"sync"

	"privacycheck/interfaces"
)

// Processor 文件处理器实现
type Processor struct {
	cache    map[string]*fileutils.FileInfo // 文件信息缓存
	cacheMux sync.RWMutex
}

// NewProcessor 创建新的文件处理器
func NewProcessor() interfaces.FileProcessor {
	return &Processor{
		cache: make(map[string]*fileutils.FileInfo),
	}
}

// GetFiles 获取文件列表
func (p *Processor) GetFiles(target string, excludeExt []string, limitSize int64) ([]string, error) {
	var files []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 检查目标是否为文件
	info, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("无法访问目标路径: %w", err)
	}

	if !info.IsDir() {
		// 单个文件
		if !p.ShouldExcludeFile(target, excludeExt, limitSize) {
			return []string{target}, nil
		}
		return []string{}, nil
	}

	// 目录遍历
	err = filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logging.Warnf("访问文件失败 %s: %v", path, err)
			return nil // 继续处理其他文件
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 并发检查文件
		wg.Add(1)
		go func(filePath string, fileInfo os.FileInfo) {
			defer wg.Done()

			if !p.ShouldExcludeFile(filePath, excludeExt, limitSize) {
				mu.Lock()
				files = append(files, filePath)
				mu.Unlock()
			}
		}(path, info)

		return nil
	})

	wg.Wait()

	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %w", err)
	}

	logging.Infof("文件扫描完成: 发现 %d 个有效文件", len(files))
	return files, nil
}

// ShouldExcludeFile 检查是否应该排除文件
func (p *Processor) ShouldExcludeFile(filePath string, excludeExt []string, limitSize int64) bool {
	// 检查缓存
	p.cacheMux.RLock()
	if fileInfo, exists := p.cache[filePath]; exists {
		p.cacheMux.RUnlock()
		return p.shouldExcludeByInfo(fileInfo, excludeExt, limitSize)
	}
	p.cacheMux.RUnlock()

	// 获取文件信息
	fileInfo, err := p.getFileInfo(filePath)
	if err != nil {
		logging.Warnf("获取文件信息失败 %s: %v", filePath, err)
		return true // 出错时排除
	}

	// 缓存文件信息
	p.cacheMux.Lock()
	p.cache[filePath] = fileInfo
	p.cacheMux.Unlock()

	return p.shouldExcludeByInfo(fileInfo, excludeExt, limitSize)
}

// shouldExcludeByInfo 根据文件信息判断是否排除
func (p *Processor) shouldExcludeByInfo(fileInfo *fileutils.FileInfo, excludeExt []string, limitSize int64) bool {
	// 检查文件大小
	if limitSize > 0 && fileInfo.Size > limitSize {
		return true
	}

	// 检查文件扩展名
	if len(excludeExt) > 0 {
		ext := strings.ToLower(filepath.Ext(fileInfo.Path))
		if ext != "" {
			ext = ext[1:] // 移除点号
		}

		for _, excludeExtension := range excludeExt {
			if strings.ToLower(excludeExtension) == ext {
				return true
			}
		}
	}

	return false
}

// getFileInfo 获取文件信息
func (p *Processor) getFileInfo(filePath string) (*fileutils.FileInfo, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// 检测文件编码
	encoding, err := p.DetectFileEncoding(filePath)
	if err != nil {
		logging.Warnf("检测文件编码失败 %s: %v", filePath, err)
		encoding = "utf-8" // 默认编码
	}

	return &fileutils.FileInfo{
		Path:     filePath,
		Size:     stat.Size(),
		Encoding: encoding,
	}, nil
}

// DetectFileEncoding 检测文件编码
func (p *Processor) DetectFileEncoding(filePath string) (string, error) {
	return fileutils.DetectFileEncoding(filePath)
}

// ClearCache 清理文件信息缓存
func (p *Processor) ClearCache() {
	p.cacheMux.Lock()
	defer p.cacheMux.Unlock()

	p.cache = make(map[string]*fileutils.FileInfo)
}

// GetCacheSize 获取缓存大小
func (p *Processor) GetCacheSize() int {
	p.cacheMux.RLock()
	defer p.cacheMux.RUnlock()

	return len(p.cache)
}
