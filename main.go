package main

import (
	"errors"
	"fmt"
	"os"
	"privacycheck/baserule"
	"privacycheck/output"
	"privacycheck/pkg/fileutils"
	"privacycheck/pkg/logging"
	"privacycheck/scanner"
	"runtime"

	"github.com/jessevdk/go-flags"
)

// CmdConfig 表示程序配置
type CmdConfig struct {
	// 基础配置
	RulesFile string `short:"r" long:"rules" description:"规则文件路径 (默认: config.yaml)" default:"config.yaml"`

	ProjectPath string `short:"p" long:"project-path" description:"项目路径，待扫描的目标文件或目录路径[必需] "`
	ProjectName string `short:"n" long:"project-name" description:"项目名称，用于生成输出文件名和缓存文件名 (默认: default_project)" default:"default_project"`

	// 过滤文件
	ExcludePath []string `long:"ep" description:"排除的路径关键字列表 (支持多个关键字 如: /tmp,/cache)"`
	ExcludeExt  []string `long:"ee" description:"排除的文件扩展名列表 (支持多个关键字 如: .tmp,.log,.bak ...)"`
	LimitSize   int      `long:"ls" description:"文件大小限制 单位:MB (超过此大小时使用被过滤, 0表示无限制, 默认: 5)" default:"5"`
	LimitChunk  int      `long:"lc" description:"分块读取阈值 单位:MB (超过此大小时使用分块读取, 0表示禁用, 默认: 5)" default:"5"`

	// 读取配置
	SaveCache bool `short:"s" long:"save-cache" description:"启用扫描结果缓存, 支持断点续扫, 推荐大项目使用"`

	// 筛选规则
	FilterNames   []string `short:"N" long:"filter-names" description:"按规则名称关键字过滤 (支持多个关键字)"`
	FilterGroups  []string `short:"G" long:"filter-groups" description:"按规则组名称关键字过滤 (支持多个关键字)"`
	SensitiveOnly bool     `short:"S" long:"sensitive-only" description:"仅启用标记为敏感信息的规则 (sensitive: true)"`

	// 性能配置
	Workers int `short:"w" long:"workers" description:"并发工作线程数 (默认: 8)" default:"8"`

	// 输出配置
	OutputFile    string   `short:"o" long:"output-file" description:"输出文件路径 (默认: {项目名称}.{格式})"`
	OutputGroup   bool     `short:"g" long:"output-group" description:"按规则组分别输出到不同文件"`
	OutputKeys    []string `short:"O" long:"output-keys" description:"指定输出字段 (可选: file,group,rule_name,match,context,position,line_number,sensitive)"`
	OutputFormat  string   `short:"f" long:"output-format" description:"输出文件格式" choice:"json" choice:"csv" default:"json"`
	FormatResults bool     `short:"F" long:"format-results" description:"格式化输出结果，清理多余的引号和空格 (默认: 启用)"`
	BlockMatches  []string `short:"b" long:"block-matches" description:"匹配结果黑名单过滤关键字列表"`

	// 日志配置
	LogFile   string `long:"lf" description:"日志文件 (为空则不写入文件)" default:""`
	LogLevel  string `long:"ll" description:"日志级别 (debug/info/warn/error)" choice:"debug" choice:"info" choice:"warn" choice:"error" default:"info"`
	LogFormat string `long:"cf" description:"控制台日志格式 (T=时间,L=级别,C=调用者,M=消息,F=函数,off=关闭)" default:"TLM"`
}

// ParseArgs 解析命令行参数
func ParseArgs() (*CmdConfig, error) {
	var opts CmdConfig

	// 设置默认值
	opts.Workers = runtime.NumCPU()
	opts.FormatResults = true

	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "PrivacyCheck - High-performance sensitive information check tool"

	// 解析命令行参数
	_, err := parser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) {
			if errors.Is(flagsErr.Type, flags.ErrHelp) {
				return nil, nil // 显示帮助信息后退出
			}
		}
		return nil, fmt.Errorf("failed to parse command-line parameters: %w", err)
	}

	// 验证参数
	if err := checkArgs(&opts); err != nil {
		return nil, err
	}

	return &opts, nil
}

// checkArgs 验证配置参数
func checkArgs(opts *CmdConfig) error {
	// 验证工作线程数
	if opts.Workers <= 0 {
		opts.Workers = runtime.NumCPU()
	}

	// 验证必需参数
	if opts.ProjectPath == "" {
		return fmt.Errorf("the project path must be specified")
	}

	// 验证文件/目录是否存在
	if opts.ProjectPath != "" {
		if exists, _, _ := fileutils.PathExists(opts.ProjectPath); !exists {
			return fmt.Errorf("the project path not exist: %s", opts.ProjectPath)
		}
	}

	// 设置默认项目名称
	if opts.ProjectName == "" {
		opts.ProjectName = fileutils.GetPathLastDir(opts.ProjectPath)
	}

	// 设置默认输出文件名
	if opts.OutputFile == "" {
		opts.OutputFile = opts.ProjectName + "." + opts.OutputFormat
	}

	// 验证输出键的有效性
	if len(opts.OutputKeys) > 0 {
		allowedKeys := map[string]bool{
			"file":        true,
			"group":       true,
			"rule_name":   true,
			"match":       true,
			"context":     true,
			"position":    true,
			"line_number": true,
			"sensitive":   true,
		}

		for _, key := range opts.OutputKeys {
			if !allowedKeys[key] {
				return fmt.Errorf("invalid output key: %s, allowed keys: file, group, rule_name, match, context, position, line_number, sensitive", key)
			}
		}
	}
	return nil
}

func main() {
	// 解析命令行参数
	cmdConfig, err := ParseArgs()
	if err != nil {
		fmt.Printf("Parameter parsing failed: %v\n", err)
		os.Exit(1)
	}

	// 如果只是显示帮助信息，直接退出
	if cmdConfig == nil {
		return
	}

	// 初始化日志记录器
	logCfg := logging.NewLogConfig(cmdConfig.LogLevel, cmdConfig.LogFile, cmdConfig.LogFormat)
	if err := logging.InitLogger(logCfg); err != nil {
		// 这里不能使用logging，因为还没初始化
		fmt.Printf("init logger failed: %v\n", err)
		os.Exit(1)
	}
	defer logging.Sync()

	logging.Infof("project name: %sproject path: %s", cmdConfig.ProjectName, cmdConfig.ProjectPath)

	// 如果配置文件不存在，创建默认配置文件
	if !fileutils.FileExists(cmdConfig.RulesFile) {
		logging.Warnf("config file %s not exist, will create default config ...", cmdConfig.RulesFile)
		if err := baserule.CreateDefaultConfig(cmdConfig.RulesFile); err != nil {
			logging.Errorf("create default config %s error:%v", cmdConfig.RulesFile, err)
		}
		logging.Infof("default config file has been created: %s", cmdConfig.RulesFile)
		os.Exit(0)
	}

	// 加载规则配置
	rulesConfig, err := baserule.LoadRulesYaml(cmdConfig.RulesFile)
	if err != nil {
		logging.Errorf("Loading the rule config failed: %v", err)
		os.Exit(1)
	}

	if len(rulesConfig.Rules) == 0 {
		logging.Errorf("rule config is empty, please check your rules")
		os.Exit(1)
	}
	logging.Infof("Load rule file rules group: %d", len(rulesConfig.Rules))

	// 验证规则
	if err := rulesConfig.ValidateRules(); err != nil {
		logging.Errorf("rule content validation failed: %v", err)
		os.Exit(1)
	}

	// 过滤规则
	filteredRules := rulesConfig.FilterRules(cmdConfig.FilterGroups, cmdConfig.FilterNames, cmdConfig.SensitiveOnly)
	ruleCount := filteredRules.CountRules()

	logging.Infof("filtered rules group: %d, rules count:%d", len(filteredRules), ruleCount)

	if ruleCount == 0 {
		logging.Error("No rules be selected. Please check the filter conditions.")
		os.Exit(1)
	}

	// 打印规则信息
	filteredRules.PrintRulesInfo()

	// 创建扫描器
	scannerConfig := newScannerConfig(cmdConfig)
	scannerInstance, err := scanner.NewScanner(filteredRules, scannerConfig)
	if err != nil {
		logging.Errorf("创建扫描器失败: %v", err)
		os.Exit(1)
	}

	// 获取待扫描文件 - 使用 fileutils 直接进行过滤
	files, err := fileutils.GetFilesWithFilter(
		cmdConfig.ProjectPath,
		cmdConfig.ExcludeExt,
		cmdConfig.ExcludePath,
		cmdConfig.LimitSize,
	)
	if err != nil {
		logging.Errorf("failed to get files with filter: %v", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		logging.Warn("no files meeting the requirements were found.")
		os.Exit(0)
	}

	logging.Infof("found %d files to be scanned", len(files))

	// 执行扫描
	results, err := scannerInstance.Scan(files)
	if err != nil {
		logging.Errorf("扫描失败: %v", err)
		os.Exit(1)
	}

	// 获取扫描统计信息
	stats := scannerInstance.GetStats()
	logging.Infof("扫描完成，发现 %d 个结果", len(results))

	// 处理输出
	if len(results) > 0 {
		outputProcessor := newOutputConfig(cmdConfig)
		if err := outputProcessor.ProcessResults(results, stats); err != nil {
			logging.Errorf("输出结果失败: %v", err)
			os.Exit(1)
		}
	} else {
		logging.Info("未发现任何敏感信息")
	}

	logging.Info("程序执行完成")
}

// newScannerConfig 从命令行配置创建扫描器配置
func newScannerConfig(cmdConfig *CmdConfig) *scanner.ScanConfig {
	return &scanner.ScanConfig{
		Workers:     cmdConfig.Workers,
		SaveCache:   cmdConfig.SaveCache,
		ProjectName: cmdConfig.ProjectName,
		ChunkLimit:  cmdConfig.LimitChunk,
	}
}

// newOutputConfig 从命令行配置创建输出配置
func newOutputConfig(cmdConfig *CmdConfig) *output.Output {
	return &output.Output{
		OutputFile:    cmdConfig.OutputFile,
		OutputGroup:   cmdConfig.OutputGroup,
		OutputKeys:    cmdConfig.OutputKeys,
		OutputFormat:  cmdConfig.OutputFormat,
		FormatResults: cmdConfig.FormatResults,
		BlockMatches:  cmdConfig.BlockMatches,
		ProjectName:   cmdConfig.ProjectName,
	}
}
