package config

import "runtime"

// Config 表示程序配置
type Config struct {
	// 基础配置
	RulesFile   string `short:"r" long:"rules" description:"规则文件路径 (默认: config.yaml, 支持 .yaml/.yml/.json 格式)" default:"config.yaml"`
	Target      string `short:"t" long:"target" description:"[必需] 待扫描的目标文件或目录路径"`
	ProjectName string `short:"p" long:"project-name" description:"项目名称，用于生成输出文件名和缓存文件名 (默认: default_project)" default:"default_project"`

	// 性能配置
	Workers   int  `short:"w" long:"workers" description:"并发工作线程数 (范围: 1-100, 默认: CPU核心数)" default:"0"`
	LimitSize int  `short:"l" long:"limit-size" description:"单文件大小限制，单位MB (范围: 0-1000, 0表示无限制, 默认: 5)" default:"5"`
	SaveCache bool `short:"s" long:"save-cache" description:"启用扫描结果缓存，支持断点续扫 (推荐大项目使用)"`
	ChunkMode bool `short:"k" long:"chunk-mode" description:"启用分块读取模式，降低内存占用但增加扫描时间"`

	// 过滤配置
	ExcludeExt []string `short:"e" long:"exclude-ext" description:"排除的文件扩展名列表 (如: .tmp,.log,.bak)"`

	// 筛选规则
	SensitiveOnly bool     `short:"S" long:"sensitive-only" description:"仅启用标记为敏感信息的规则 (sensitive: true)"`
	FilterNames   []string `short:"N" long:"filter-names" description:"按规则名称关键字过滤 (支持多个关键字)"`
	FilterGroups  []string `short:"G" long:"filter-groups" description:"按规则组名称关键字过滤 (支持多个关键字)"`

	// 输出配置
	OutputFile    string   `short:"o" long:"output-file" description:"输出文件路径 (默认: {项目名称}.{格式})"`
	OutputGroup   bool     `short:"g" long:"output-group" description:"按规则组分别输出到不同文件"`
	OutputKeys    []string `short:"O" long:"output-keys" description:"指定输出字段 (可选: file,group,rule_name,match,context,position,line_number,sensitive)"`
	OutputFormat  string   `short:"f" long:"output-format" description:"输出文件格式" choice:"json" choice:"csv" default:"json"`
	FormatResults bool     `short:"F" long:"format-results" description:"格式化输出结果，清理多余的引号和空格 (默认: 启用)"`
	BlockMatches  []string `short:"b" long:"block-matches" description:"匹配结果黑名单过滤关键字列表"`

	// 辅助工具
	Version bool `short:"v" long:"version" description:"显示版本信息并退出"`

	// 日志配置
	LogLevel      string `long:"log-level" description:"日志级别" choice:"debug" choice:"info" choice:"warn" choice:"error" default:"info"`
	LogFile       string `long:"log-file" description:"日志文件路径 (为空则不写入文件)"`
	ConsoleFormat string `long:"console-format" description:"控制台日志格式 (T=时间,L=级别,C=调用者,M=消息,F=函数, 'off'=关闭控制台输出)" default:"TLM"`
}

// GetDefaultWorkers 获取默认工作线程数
func GetDefaultWorkers() int {
	return runtime.NumCPU()
}
