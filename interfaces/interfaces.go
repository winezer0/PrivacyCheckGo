package interfaces

import (
	"privacycheck/core"
	"privacycheck/config"
)

// Scanner 扫描器接口
type Scanner interface {
	Scan(files []string, config *config.Config) ([]core.ScanResult, *core.ScanStats, error)
	GetStats() *core.ScanStats
}

// RuleEngine 规则引擎接口
type RuleEngine interface {
	LoadRules(rules []config.Rule) error
	ApplyRules(content string, filePath string) []core.ScanResult
	GetRulesCount() int
	GetGroupsCount() int
}

// FileProcessor 文件处理器接口
type FileProcessor interface {
	GetFiles(target string, excludeExt []string, limitSize int64) ([]string, error)
	ShouldExcludeFile(filePath string, excludeExt []string, limitSize int64) bool
	DetectFileEncoding(filePath string) (string, error)
}

// OutputProcessor 输出处理器接口
type OutputProcessor interface {
	ProcessResults(results []core.ScanResult, stats *core.ScanStats, config *config.Config) error
	FormatResults(results []core.ScanResult, format string) ([]byte, error)
}

// CacheManager 缓存管理器接口
type CacheManager interface {
	LoadCache(filePath string) (*core.ScanCached, error)
	SaveCache(cache *core.ScanCached, filePath string) error
	GetCachedResult(filePath string) ([]core.ScanResult, bool)
	SetCachedResult(filePath string, results []core.ScanResult)
}

// Logger 日志接口
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}