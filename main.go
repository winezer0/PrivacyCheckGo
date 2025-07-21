package main

import (
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"privacycheck/pkg/logging"
	"runtime"
	"strings"

	"privacycheck/config"
	"privacycheck/fileutils"
	"privacycheck/output"
	"privacycheck/scanner"
)

// CmdConfig 表示程序配置
type CmdConfig struct {
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

// ParseArgs 解析命令行参数
func ParseArgs() (*CmdConfig, error) {
	var opts CmdConfig

	// 设置默认值
	opts.Workers = runtime.NumCPU()
	opts.FormatResults = true

	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "PrivacyCheck - 高性能隐私信息检测工具\n\n使用示例:\n  privacycheck -t /path/to/project\n  privacycheck -t /path/to/project -w 8 -s\n  privacycheck -t /path/to/project -S -f csv"

	// 解析命令行参数
	_, err := parser.Parse()
	if err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) {
			if errors.Is(flagsErr.Type, flags.ErrHelp) {
				return nil, nil // 显示帮助信息后退出
			}
		}
		return nil, fmt.Errorf("解析命令行参数失败: %w", err)
	}

	// 处理版本信息显示
	if opts.Version {
		fmt.Println(config.GetVersionInfo())
		return nil, nil
	}

	// 验证参数
	if err := validateConfig(&opts); err != nil {
		return nil, err
	}

	return &opts, nil
}

// validateConfig 验证配置参数
func validateConfig(opts *CmdConfig) error {
	// 验证必需参数
	if opts.Target == "" && !opts.Version {
		return fmt.Errorf("必须指定 --target 参数")
	}

	// 验证文件/目录是否存在
	if opts.Target != "" {
		if _, err := os.Stat(opts.Target); os.IsNotExist(err) {
			return fmt.Errorf("目标路径不存在: %s", opts.Target)
		}
	}

	// 验证工作线程数
	if opts.Workers <= 0 {
		opts.Workers = runtime.NumCPU()
	} else if opts.Workers > 100 {
		return fmt.Errorf("工作线程数不能超过100，当前值: %d", opts.Workers)
	}

	// 验证文件大小限制
	if opts.LimitSize < 0 {
		return fmt.Errorf("文件大小限制不能为负数，当前值: %d", opts.LimitSize)
	} else if opts.LimitSize > 1000 {
		return fmt.Errorf("文件大小限制不能超过1000MB，当前值: %d", opts.LimitSize)
	}

	// 验证输出格式
	if opts.OutputFormat != "json" && opts.OutputFormat != "csv" {
		return fmt.Errorf("不支持的输出格式: %s，支持的格式: json, csv", opts.OutputFormat)
	}

	// 验证日志级别
	validLogLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLogLevels {
		if opts.LogLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("无效的日志级别: %s，支持的级别: %s", opts.LogLevel, strings.Join(validLogLevels, ", "))
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
				return fmt.Errorf("无效的输出键: %s，允许的键: file, group, rule_name, match, context, position, line_number, sensitive", key)
			}
		}
	}

	// 设置默认输出文件名
	if opts.OutputFile == "" {
		opts.OutputFile = opts.ProjectName + "." + opts.OutputFormat
	}

	return nil
}

func main() {
	// 解析命令行参数
	cmdConfig, err := ParseArgs()
	if err != nil {
		fmt.Printf("参数解析失败: %v\n", err)
		os.Exit(1)
	}

	// 如果只是显示帮助信息，直接退出
	if cmdConfig == nil {
		return
	}

	// 初始化日志记录器
	logCfg := logging.NewLogConfig(cmdConfig.LogLevel, cmdConfig.LogFile, cmdConfig.ConsoleFormat)
	if err := logging.InitLogger(logCfg); err != nil {
		// 这里不能使用logging，因为还没初始化
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logging.Sync()

	logging.Info("PrivacyCheck Go版本启动")
	logging.Infof("目标路径: %s", cmdConfig.Target)
	logging.Infof("项目名称: %s", cmdConfig.ProjectName)

	// 加载规则配置
	rulesConfig, err := config.LoadRulesConfig(cmdConfig.RulesFile)
	if err != nil {
		logging.Errorf("加载规则配置失败: %v", err)
		os.Exit(1)
	}

	logging.Infof("加载规则配置: %d 个规则组", len(rulesConfig.Rules))

	// 验证规则
	if err := rulesConfig.ValidateRules(); err != nil {
		logging.Errorf("规则验证失败: %v", err)
		os.Exit(1)
	}

	// 过滤规则
	filteredRules := rulesConfig.FilterRules(cmdConfig.FilterGroups, cmdConfig.FilterNames, cmdConfig.SensitiveOnly)
	ruleCount := filteredRules.CountRules()

	logging.Infof("过滤后规则: %d 个规则组, %d 个规则", len(filteredRules), ruleCount)

	if ruleCount == 0 {
		logging.Error("没有可用的规则，请检查过滤条件")
		os.Exit(1)
	}

	// 打印规则信息
	filteredRules.PrintRulesInfo()

	// 获取待扫描文件
	files, err := fileutils.GetFilesWithFilter(cmdConfig.Target, cmdConfig.ExcludeExt, cmdConfig.LimitSize)
	if err != nil {
		logging.Errorf("获取文件列表失败: %v", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		logging.Warn("没有找到符合条件的文件")
		return
	}

	logging.Infof("发现 %d 个待扫描文件", len(files))

	// 创建扫描器
	scannerInstance, err := scanner.NewScanner(filteredRules, cmdConfig)
	if err != nil {
		logging.Errorf("创建扫描器失败: %v", err)
		os.Exit(1)
	}

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
		outputProcessor := output.NewOutput(cmdConfig)
		if err := outputProcessor.ProcessResults(results, stats); err != nil {
			logging.Errorf("输出结果失败: %v", err)
			os.Exit(1)
		}
	} else {
		logging.Info("未发现任何敏感信息")
	}

	logging.Info("程序执行完成")
}
