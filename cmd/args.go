package cmd

import (
	"fmt"
	"os"
	"runtime"

	"privacycheck/core"

	"github.com/jessevdk/go-flags"
)

// ParseArgs 解析命令行参数
func ParseArgs() (*core.Config, error) {
	var config core.Config

	// 设置默认值
	config.Workers = runtime.NumCPU()
	config.FormatResults = true

	parser := flags.NewParser(&config, flags.Default)
	parser.Usage = "Privacy information detection tool"

	// 解析命令行参数
	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok {
			if flagsErr.Type == flags.ErrHelp {
				return nil, nil // 显示帮助信息后退出
			}
		}
		return nil, fmt.Errorf("解析命令行参数失败: %w", err)
	}

	// 处理版本信息显示
	if config.Version {
		fmt.Println(core.GetVersionString())
		return nil, nil
	}

	// 验证必需参数（除非是版本命令）
	if config.Target == "" && !config.Version {
		return nil, fmt.Errorf("必须指定 --target 参数")
	}

	// 验证文件/目录是否存在
	if _, err := os.Stat(config.Target); os.IsNotExist(err) {
		return nil, fmt.Errorf("目标路径不存在: %s", config.Target)
	}

	// 设置默认工作线程数
	if config.Workers <= 0 {
		config.Workers = runtime.NumCPU()
	}

	// 验证输出键的有效性
	if len(config.OutputKeys) > 0 {
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

		for _, key := range config.OutputKeys {
			if !allowedKeys[key] {
				return nil, fmt.Errorf("无效的输出键: %s，允许的键: file, group, rule_name, match, context, position, line_number, sensitive", key)
			}
		}
	}

	// 设置默认输出文件名
	if config.OutputFile == "" {
		config.OutputFile = config.ProjectName + "." + config.OutputFormat
	}

	return &config, nil
}

// PrintVersion 打印版本信息
func PrintVersion() {
	fmt.Println("PrivacyCheck Go版本 v1.0.0")
	fmt.Println("基于Python版本PrivacyCheck重新实现的Go版本")
	fmt.Println("兼容HAE规则格式的静态代码敏感信息检测工具")
}
