package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	"privacycheck/config"
)

// ParseArgs 解析命令行参数
func ParseArgs() (*config.Config, error) {
	var opts config.Config

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
func validateConfig(opts *config.Config) error {
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