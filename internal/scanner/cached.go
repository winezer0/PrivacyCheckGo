package scanner

import (
	"encoding/json"
	"fmt"
	"github.com/winezer0/xutils/logging"
	"github.com/winezer0/xutils/utils"
	"os"
	"sync"
	"time"
)

// CacheData 表示扫描缓存
type CacheData struct {
	Result     map[string][]ScanResult `json:"result"`      // 缓存的扫描结果
	LastUpdate string                  `json:"last_update"` // 最后更新时间
}

// CacheManager 缓存管理器实现
type CacheManager struct {
	cacheFile string
	cacheData *CacheData
	cacheMux  sync.RWMutex
	lastSave  time.Time
	dirty     bool // 标记是否有未保存的更改
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(filePath string) *CacheManager {
	m := &CacheManager{
		cacheFile: filePath,
		cacheData: &CacheData{
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
	m.cacheMux.Lock()
	defer m.cacheMux.Unlock()

	if _, err := os.Stat(m.cacheFile); os.IsNotExist(err) {
		return nil // 缓存文件不存在，不是错误
	}

	data, err := utils.ReadFileToBytes(m.cacheFile)
	if err != nil {
		return fmt.Errorf("failed to read cache file [%s]: %w", m.cacheFile, err)
	}

	if err := json.Unmarshal(data, m.cacheData); err != nil {
		return fmt.Errorf("failed to parse cache file [%s]: %w", m.cacheFile, err)
	}

	logging.Infof("cache file [%s] loaded successfully: cached results: %d, last update: %s", m.cacheFile, len(m.cacheData.Result), m.cacheData.LastUpdate)

	return nil
}

// SaveCache 保存缓存文件
func (m *CacheManager) SaveCache() error {
	m.cacheMux.Lock()
	defer m.cacheMux.Unlock()

	// 调用了保存功能就应该更新时间戳
	m.lastSave = time.Now()
	m.cacheData.LastUpdate = time.Now().Format(time.RFC3339)
	if err := utils.SaveJSON(m.cacheFile, m.cacheData); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}
	m.dirty = false
	return nil
}

// GetCachedResult 获取缓存结果
func (m *CacheManager) GetCachedResult(filePath string) ([]ScanResult, bool) {
	m.cacheMux.RLock()
	defer m.cacheMux.RUnlock()

	results, exists := m.cacheData.Result[filePath]
	return results, exists
}

// SetCachedResult 设置缓存结果
func (m *CacheManager) SetCachedResult(filePath string, results []ScanResult) {
	m.cacheMux.Lock()
	defer m.cacheMux.Unlock()

	m.cacheData.Result[filePath] = results
	m.dirty = true
}

// ShouldSave 检查是否应该保存缓存
func (m *CacheManager) ShouldSave() bool {
	m.cacheMux.RLock()
	defer m.cacheMux.RUnlock()

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
	m.cacheMux.RLock()
	defer m.cacheMux.RUnlock()

	return len(m.cacheData.Result), m.cacheData.LastUpdate
}
