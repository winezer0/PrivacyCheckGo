package scanner

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
