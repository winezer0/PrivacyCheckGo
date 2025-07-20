package core

import (
	"time"
)

// Rule 表示单个检测规则
type Rule struct {
	Name         string `yaml:"name" json:"name"`                             // 规则名称
	FRegex       string `yaml:"f_regex" json:"f_regex"`                       // 正则表达式
	SRegex       string `yaml:"s_regex,omitempty" json:"s_regex,omitempty"`   // 二次匹配规则(未实现)
	Format       string `yaml:"format,omitempty" json:"format,omitempty"`     // 结果提取格式(未实现)
	Color        string `yaml:"color,omitempty" json:"color,omitempty"`       // 结果颜色显示(未实现)
	Scope        string `yaml:"scope,omitempty" json:"scope,omitempty"`       // 规则匹配范围(未实现)
	Engine       string `yaml:"engine,omitempty" json:"engine,omitempty"`     // 规则匹配引擎(未实现)
	Sensitive    bool   `yaml:"sensitive" json:"sensitive"`                   // 是否敏感信息
	Loaded       bool   `yaml:"loaded" json:"loaded"`                         // 是否启用规则
	IgnoreCase   bool   `yaml:"ignore_case" json:"ignore_case"`               // 正则匹配时忽略大小写
	ContextLeft  int    `yaml:"context_left,omitempty" json:"context_left"`   // 匹配结果向左扩充字符数
	ContextRight int    `yaml:"context_right,omitempty" json:"context_right"` // 匹配结果向右扩充字符数
}

// RuleGroup 表示规则组
type RuleGroup struct {
	Group string `yaml:"group" json:"group"` // 规则组名称
	Rule  []Rule `yaml:"rule" json:"rule"`   // 规则列表
}

// RulesConfig 表示完整的规则配置
type RulesConfig struct {
	Rules []RuleGroup `yaml:"rules" json:"rules"`
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

// ScanCache 表示扫描缓存
type ScanCache struct {
	Result     map[string][]ScanResult `json:"result"`      // 缓存的扫描结果
	LastUpdate string                  `json:"last_update"` // 最后更新时间
}

// Config 表示程序配置
type Config struct {
	// 基础配置
	RulesFile   string `short:"r" long:"rules" description:"规则文件的路径" default:"config.yaml"`
	Target      string `short:"t" long:"target" description:"待扫描的项目目标文件或目录"`
	ProjectName string `short:"p" long:"project-name" description:"项目名称, 影响默认输出文件名和缓存文件名" default:"default_project"`

	// 性能配置
	Workers   int  `short:"w" long:"workers" description:"工作线程数量(默认值：CPU 核心数)" default:"0"`
	LimitSize int  `short:"l" long:"limit-size" description:"检查文件大小限制 不超过 limit_size M" default:"5"`
	SaveCache bool `short:"s" long:"save-cache" description:"定时缓存扫描结果, 建议大项目使用"`
	ChunkMode bool `short:"k" long:"chunk-mode" description:"使用chunk模式读取文件,运行时间延长,内存占用减小"`

	// 过滤配置
	ExcludeExt []string `short:"e" long:"exclude-ext" description:"排除文件扩展名"`

	// 筛选规则
	SensitiveOnly bool     `short:"S" long:"sensitive-only" description:"只启用敏感信息规则 (sensitive: true)"`
	FilterNames   []string `short:"N" long:"filter-names" description:"仅启用name中包含指定关键字的规则"`
	FilterGroups  []string `short:"G" long:"filter-groups" description:"仅启用group中包含指定关键字的规则"`

	// 输出配置
	OutputFile    string   `short:"o" long:"output-file" description:"指定输出文件路径"`
	OutputGroup   bool     `short:"g" long:"output-group" description:"为规则组单独输出结果"`
	OutputKeys    []string `short:"O" long:"output-keys" description:"仅输出结果中指定键的值"`
	OutputFormat  string   `short:"f" long:"output-format" description:"指定输出文件格式: json 或 csv" choice:"json" choice:"csv" default:"json"`
	FormatResults bool     `short:"F" long:"format-results" description:"对输出结果的每个值进行格式化，去除引号、空格等符号"`
	BlockMatches  []string `short:"b" long:"block-matches" description:"对匹配结果中的match键值进行黑名单关键字列表匹配剔除"`

	// 辅助工具
	Version bool `short:"v" long:"version" description:"显示版本信息"`

	// 日志配置
	LogLevel      string `long:"log-level" description:"日志级别" choice:"debug" choice:"info" choice:"warn" choice:"error" default:"info"`
	LogFile       string `long:"log-file" description:"日志文件路径（为空则不写入文件）"`
	ConsoleFormat string `long:"console-format" description:"控制台格式字符串，空或 off 表示关闭控制台输出" default:"TLM"`
}

// FileInfo 表示文件信息
type FileInfo struct {
	Path     string
	Size     int64
	Encoding string
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
