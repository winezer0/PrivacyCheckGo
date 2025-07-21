package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"privacycheck/pkg/fileutils"
	"privacycheck/pkg/logging"
	"sync"
	"time"
)

// CacheManager 缓存管理器实现
type CacheManager struct {
	data     *ScanCached
	filePath string
	mux      sync.RWMutex
	lastSave time.Time
	dirty    bool // 标记是否有未保存的更改
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(filePath string) *CacheManager {
	m := &CacheManager{
		filePath: filePath,
		data: &ScanCached{
			Result:     make(map[string][]ScanResult),
			LastUpdate: time.Now().Format(time.RFC3339),
		},
		lastSave: time.Now(),
		dirty:    false,
	}

	// 尝试加载现有缓存
	if err := m.LoadCache(); err != nil {
		logging.Warnf("加载缓存失败: %v", err)
	}

	return m
}

// LoadCache 加载缓存文件
func (m *CacheManager) LoadCache() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		return nil // 缓存文件不存在，不是错误
	}

	data, err := fileutils.ReadFile(m.filePath)
	if err != nil {
		return fmt.Errorf("读取缓存文件失败: %w", err)
	}

	if err := json.Unmarshal(data, m.data); err != nil {
		return fmt.Errorf("解析缓存文件失败: %w", err)
	}

	logging.Infof("加载缓存完成: 已缓存结果数: %d, 缓存更新时间: %s",
		len(m.data.Result), m.data.LastUpdate)

	return nil
}

// SaveCache 保存缓存文件
func (m *CacheManager) SaveCache() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	// 调用了保存功能就应该更新时间戳
	m.data.LastUpdate = time.Now().Format(time.RFC3339)

	if err := fileutils.WriteJSONFile(m.filePath, m.data); err != nil {
		return fmt.Errorf("写入缓存文件失败: %w", err)
	}

	m.lastSave = time.Now()
	m.dirty = false
	return nil
}

// GetCachedResult 获取缓存结果
func (m *CacheManager) GetCachedResult(filePath string) ([]ScanResult, bool) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	results, exists := m.data.Result[filePath]
	return results, exists
}

// SetCachedResult 设置缓存结果
func (m *CacheManager) SetCachedResult(filePath string, results []ScanResult) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.data.Result[filePath] = results
	m.dirty = true
}

// ShouldSave 检查是否应该保存缓存
func (m *CacheManager) ShouldSave() bool {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return m.dirty && time.Since(m.lastSave) >= 10*time.Second
}

// AutoSave 自动保存（如果需要）
func (m *CacheManager) AutoSave() error {
	if m.ShouldSave() {
		return m.SaveCache()
	}
	return nil
}

// ForceSave 强制保存
func (m *CacheManager) ForceSave() error {
	return m.SaveCache()
}

// GetCacheStats 获取缓存统计信息
func (m *CacheManager) GetCacheStats() (int, string) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return len(m.data.Result), m.data.LastUpdate
}
