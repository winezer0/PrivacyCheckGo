package scanner

import "time"

// Config 扫描器配置
type Config struct {
	Workers     int
	SaveCache   bool
	ProjectName string
}

// ScanResult 表示扫描结果
type ScanResult struct {
	File       string `json:"file"`        // 文件路径
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

// ScanCached 表示扫描缓存
type ScanCached struct {
	Result     map[string][]ScanResult `json:"result"`      // 缓存的扫描结果
	LastUpdate string                  `json:"last_update"` // 最后更新时间
}
