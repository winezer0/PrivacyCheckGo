package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"privacycheck/pkg/logging"
	"sync"
	"time"

	"privacycheck/core"
	"privacycheck/interfaces"
)

// Manager 缓存管理器实现
type Manager struct {
	data     *core.ScanCached
	filePath string
	mux      sync.RWMutex
	lastSave time.Time
	dirty    bool // 标记是否有未保存的更改
}

// NewManager 创建新的缓存管理器
func NewManager(filePath string) interfaces.CacheManager {
	m := &Manager{
		filePath: filePath,
		data: &core.ScanCached{
			Result:     make(map[string][]core.ScanResult),
			LastUpdate: time.Now().Format(time.RFC3339),
		},
		lastSave: time.Now(),
		dirty:    false,
	}

	// 尝试加载现有缓存
	if err := m.LoadCache(filePath); err != nil {
		logging.Warnf("加载缓存失败: %v", err)
	}

	return m
}

// LoadCache 加载缓存文件
func (m *Manager) LoadCache(filePath string) (*core.ScanCached, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return m.data, nil // 缓存文件不存在，不是错误
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取缓存文件失败: %w", err)
	}

	if err := json.Unmarshal(data, m.data); err != nil {
		return nil, fmt.Errorf("解析缓存文件失败: %w", err)
	}

	logging.Infof("加载缓存完成: 已缓存结果数: %d, 缓存更新时间: %s",
		len(m.data.Result), m.data.LastUpdate)

	return m.data, nil
}

// SaveCache 保存缓存文件
func (m *Manager) SaveCache(cache *core.ScanCached, filePath string) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	if cache != nil {
		m.data = cache
	}

	m.data.LastUpdate = time.Now().Format(time.RFC3339)

	data, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入缓存文件失败: %w", err)
	}

	m.lastSave = time.Now()
	m.dirty = false
	return nil
}

// GetCachedResult 获取缓存结果
func (m *Manager) GetCachedResult(filePath string) ([]core.ScanResult, bool) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	results, exists := m.data.Result[filePath]
	return results, exists
}

// SetCachedResult 设置缓存结果
func (m *Manager) SetCachedResult(filePath string, results []core.ScanResult) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.data.Result[filePath] = results
	m.dirty = true
}

// ShouldSave 检查是否应该保存缓存
func (m *Manager) ShouldSave(forceStore bool) bool {
	m.mux.RLock()
	defer m.mux.RUnlock()

	if forceStore || !m.dirty {
		return forceStore
	}
	return time.Since(m.lastSave) >= 10*time.Second // 每10秒保存一次
}

// AutoSave 自动保存（如果需要）
func (m *Manager) AutoSave() error {
	if m.ShouldSave(false) {
		return m.SaveCache(nil, m.filePath)
	}
	return nil
}

// ForceSave 强制保存
func (m *Manager) ForceSave() error {
	return m.SaveCache(nil, m.filePath)
}

// GetCacheStats 获取缓存统计信息
func (m *Manager) GetCacheStats() (int, string) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return len(m.data.Result), m.data.LastUpdate
}
