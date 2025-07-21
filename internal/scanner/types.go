package scanner

import "time"

// ScanConfig 扫描器配置
type ScanConfig struct {
	Workers     int
	SaveCache   bool
	ProjectName string
	ChunkLimit  int // 分块读取阈值，单位MB
}

// ScanJob 扫描任务结果
type ScanJob struct {
	FilePath string
	Results  []ScanResult
	Error    error
}

// ScanResult 表示扫描结果
type ScanResult struct {
	File       string `json:"cacheFile"`   // 文件路径
	Group      string `json:"group"`       // 规则组名称
	RuleName   string `json:"rule_name"`   // 规则名称
	Match      string `json:"match"`       // 匹配的内容
	Context    string `json:"context"`     // 上下文内容
	Position   int    `json:"position"`    // 匹配位置
	LineNumber int    `json:"line_number"` // 行号
	Sensitive  bool   `json:"sensitive"`   // 是否敏感信息
}

// ScanStats 表示扫描统计信息
type ScanStats struct {
	TotalFiles     int           // 总文件数
	ProcessedFiles int           // 已处理文件数
	TotalResults   int           // 总结果数
	StartTime      time.Time     // 开始时间
	ElapsedTime    time.Duration // 已用时间
	EstimatedTime  time.Duration // 预计剩余时间
}

// ProgressInfo 表示进度信息
type ProgressInfo struct {
	Current   int
	Total     int
	Percent   float64
	Elapsed   time.Duration
	Remaining time.Duration
	Message   string
}
