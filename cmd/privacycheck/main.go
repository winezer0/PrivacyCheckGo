package main

import (
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/winezer0/xutils/logging"
	"github.com/winezer0/xutils/utils"
	"os"
	"privacycheck/internal/baserule"
	"privacycheck/internal/output"
	"privacycheck/internal/ruletest"
	"privacycheck/internal/scanner"
	"runtime"
)

const (
	AppName      = "PrivacyCheck"
	AppVersion   = "0.2.1"
	BuildDate    = "2026-02-11"
	AppShortDesc = "privacy check base on rules"
	AppLongDesc  = "privacy check base on rules"
)

// Options 表示程序配置
type Options struct {
	// 基础配置
	ProjectName string `short:"n" long:"project-name" description:"项目名称"`
	ProjectPath string `short:"p" long:"project-path" description:"扫描路径"`
	RulesFile   string `short:"r" long:"rules" description:"扫描规则文件路径" default:"config.yaml"`

	// 过滤文件
	ExcludePath []string `long:"ep" description:"排除的路径关键字列表 (支持多个关键字 如: /tmp,/cache)"`
	ExcludeExt  []string `long:"ee" description:"排除的文件扩展名列表 (支持多个关键字 如: .tmp,.log,.bak ...)"`
	LimitSize   int      `long:"ls" description:"文件大小限制 单位:MB (超过此大小时使用被过滤, 0表示无限制, 默认: 5)" default:"5"`
	LimitChunk  int      `long:"lc" description:"分块读取阈值 单位:MB (超过此大小时使用分块读取, 0表示禁用, 默认: 5)" default:"5"`

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

	// 自动化启用缓存
	Cached    bool   `long:"cached" description:"enable scan cacher"`
	scanCache string `long:"scan-cache" description:"customized scan cache file path（default <project>.<hash>.<engine>.cache）"`

	// 日志配置
	LogFile    string `long:"lf" description:"日志文件 (为空则不写入文件)" default:""`
	LogLevel   string `long:"ll" description:"日志级别 (debug/info/warn/error)" choice:"debug" choice:"info" choice:"warn" choice:"error" default:"info"`
	LogConsole string `long:"cf" description:"控制台日志格式 (T=时间,L=级别,C=调用者,M=消息,F=函数,off=关闭)" default:"TLM"`

	Test    bool `long:"test" description:"test all rules and generate test report"`
	Version bool `short:"v" long:"version" description:"display the program version and exit"`
}

func main() {
	// 初始化命令行输入配置
	opts, _ := InitOptionsArgs(1)

	// 加载规则配置
	rulesConfig, err := baserule.LoadRulesYaml(opts.RulesFile)
	if err != nil {
		logging.Fatalf("Loading the rule config failed: %v", err)
	}

	if len(rulesConfig.Rules) == 0 {
		logging.Fatalf("rule config is empty, please check your rules")
	}

	logging.Infof("Load rule file rules group: %d", len(rulesConfig.Rules))

	// 检查是否为测试模式
	if opts.Test {
		ruletest.RunRuleTest(opts.RulesFile, rulesConfig.Rules)
		return
	}

	// 验证规则
	if err := rulesConfig.ValidateRules(); err != nil {
		logging.Fatalf("rule content validation failed: %v", err)
	}

	// 过滤规则
	filteredRules := rulesConfig.FilterRules(opts.FilterGroups, opts.FilterNames, opts.SensitiveOnly)
	ruleCount := filteredRules.CountRules()

	logging.Infof("filtered rules group: %d, rules count:%d", len(filteredRules), ruleCount)

	if ruleCount == 0 {
		logging.Fatalf("No rules be selected. Please check the filter conditions.")
	}

	// 打印规则信息
	filteredRules.PrintRulesInfo()

	// 创建扫描器
	scannerConfig := &scanner.ScanConfig{
		Workers:     opts.Workers,
		ProjectName: opts.ProjectName,
		ProjectPath: opts.ProjectPath,
		CacheFile:   opts.scanCache,
		ChunkLimit:  opts.LimitChunk,
	}

	// 获取待扫描文件 - 使用 fileutils 直接进行过滤
	files, err := utils.GetFilesWithFilter(opts.ProjectPath, opts.ExcludeExt, opts.ExcludePath, opts.LimitSize)
	if err != nil || len(files) == 0 {
		logging.Fatalf("failed to get files with filter: %v", err)
	}
	logging.Infof("found %d files to be scanned", len(files))

	// 执行扫描
	instance, err := scanner.NewScanner(filteredRules, scannerConfig)
	if err != nil {
		logging.Fatalf("failed to create scanner: %v", err)
	}
	results, err := instance.Scan(files)
	if err != nil {
		logging.Fatalf("scanner scan failed: %v", err)
	}
	// 获取扫描统计信息
	logging.Infof("scan completed, found %d results", len(results))

	// 处理输出
	if len(results) > 0 {
		outputProcessor := newOutputConfig(opts)
		if err := outputProcessor.ProcessResults(results); err != nil {
			logging.Fatalf("failed to output results: %v", err)
		}
	} else {
		logging.Info("no sensitive information found")
	}
	logging.Info("program execution completed")
}

// newOutputConfig 从命令行配置创建输出配置
func newOutputConfig(cmdConfig *Options) *output.Output {
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

// InitOptionsArgs 常用的工具函数，解析parser和logging配置
func InitOptionsArgs(minimumParams int) (*Options, *flags.Parser) {
	opts := &Options{}
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = AppName
	parser.Usage = "[OPTIONS]"
	parser.ShortDescription = AppShortDesc
	parser.LongDescription = AppLongDesc

	// 命令行参数数量检查 指不包含程序名本身的参数数量
	if minimumParams > 0 && len(os.Args)-1 < minimumParams {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	// 命令行参数解析检查
	if _, err := parser.Parse(); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && errors.Is(flagsErr.Type, flags.ErrHelp) {
			os.Exit(0)
		}
		fmt.Printf("Error:%v\n", err)
		os.Exit(1)
	}

	// 版本号输出
	if opts.Version {
		fmt.Printf("%s version %s\n", AppName, AppVersion)
		fmt.Printf("Build Date: %s\n", BuildDate)
		os.Exit(0)
	}

	// 初始化日志器
	logCfg := logging.NewLogConfig(opts.LogLevel, opts.LogFile, opts.LogConsole)
	if err := logging.InitLogger(logCfg); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logging.Sync()

	// 如果配置文件不存在，创建默认配置文件
	if !utils.FileExists(opts.RulesFile) {
		logging.Warnf("config file %s not exist, will create default config ...", opts.RulesFile)
		if err := baserule.CreateDefaultConfig(opts.RulesFile); err != nil {
			logging.Errorf("create default config %s error:%v", opts.RulesFile, err)
		}
		logging.Infof("default config file has been created: %s", opts.RulesFile)
		os.Exit(0)
	}

	// 处理项目路径
	if opts.ProjectPath == "" {
		logging.Fatalf("must input project path !!!")
	}
	// 验证文件/目录是否存在
	if exists, _, _ := utils.PathExists(opts.ProjectPath); !exists {
		logging.Fatalf("project path not exist: %s", opts.ProjectPath)
	}

	// 设置默认项目名称
	if opts.ProjectName == "" {
		opts.ProjectName = utils.GetPathLastDir(opts.ProjectPath)
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
				logging.Fatalf("invalid output key: %s, allowed keys: file, group, rule_name, match, context, position, line_number, sensitive", key)
			}
		}
	}

	// 配置缓存功能
	if opts.Cached && opts.scanCache == "" {
		opts.scanCache = utils.GenProjectFileName(opts.ProjectName, opts.ProjectPath, AppName, "cache")
	}

	// 验证工作线程数
	if opts.Workers <= 0 {
		opts.Workers = utils.MaxNum(runtime.NumCPU()/4, 1)
	}

	// 检查是否为测试模式
	if opts.Test {
		if rulesConfig, err := baserule.LoadRulesYaml(opts.RulesFile); err == nil {
			ruletest.RunRuleTest(opts.RulesFile, rulesConfig.Rules)
		}
		os.Exit(0)
	}

	logging.Infof("ProjectName: %s", opts.ProjectName)
	logging.Infof("ProjectPath: %s", opts.ProjectPath)
	logging.Infof("RulesFile: %s", opts.RulesFile)
	logging.Infof("Workers: %d", opts.Workers)
	logging.Infof("Output: %s", opts.OutputFile)

	return opts, parser
}
