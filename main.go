package main

import (
	"errors"
	"fmt"
	"github.com/jessevdk/go-flags"
	"os"
	"privacycheck/output"
	"runtime"

	"privacycheck/config"
	"privacycheck/logging"
	"privacycheck/scanner"
	"privacycheck/utils"
)

// ParseArgs 解析命令行参数
func ParseArgs() (*config.Config, error) {
	var opts config.Config

	// 设置默认值
	opts.Workers = runtime.NumCPU()
	opts.FormatResults = true

	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "Privacy information detection tool"

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

	// 验证必需参数（除非是版本命令）
	if opts.Target == "" && !opts.Version {
		return nil, fmt.Errorf("必须指定 --target 参数")
	}

	// 验证文件/目录是否存在
	if _, err := os.Stat(opts.Target); os.IsNotExist(err) {
		return nil, fmt.Errorf("目标路径不存在: %s", opts.Target)
	}

	// 设置默认工作线程数
	if opts.Workers <= 0 {
		opts.Workers = runtime.NumCPU()
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
				return nil, fmt.Errorf("无效的输出键: %s，允许的键: file, group, rule_name, match, context, position, line_number, sensitive", key)
			}
		}
	}

	// 设置默认输出文件名
	if opts.OutputFile == "" {
		opts.OutputFile = opts.ProjectName + "." + opts.OutputFormat
	}

	return &opts, nil
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
	files, err := utils.GetFilesWithFilter(cmdConfig.Target, cmdConfig.ExcludeExt, cmdConfig.LimitSize)
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

	logging.Infof("扫描完成，发现 %d 个结果", len(results))

	// 处理输出
	if len(results) > 0 {
		outputProcessor := output.NewOutput(cmdConfig)
		if err := outputProcessor.ProcessResults(results); err != nil {
			logging.Errorf("输出结果失败: %v", err)
			os.Exit(1)
		}
	} else {
		logging.Info("未发现任何敏感信息")
	}

	logging.Info("程序执行完成")
}
