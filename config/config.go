package config

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
