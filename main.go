package main

import (
	"fmt"
	"os"

	"privacycheck/cmd"
	"privacycheck/config"
	"privacycheck/logging"
	"privacycheck/scanner"
	"privacycheck/utils"
)

func main() {
	// 解析命令行参数
	cmdConfig, err := cmd.ParseArgs()
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
	if err := config.ValidateRules(rulesConfig); err != nil {
		logging.Errorf("规则验证失败: %v", err)
		os.Exit(1)
	}

	// 过滤规则
	filteredRules := config.FilterRules(rulesConfig, cmdConfig.FilterGroups, cmdConfig.FilterNames, cmdConfig.SensitiveOnly)
	ruleCount := config.CountRules(filteredRules)

	logging.Infof("过滤后规则: %d 个规则组, %d 个规则", len(filteredRules), ruleCount)

	if ruleCount == 0 {
		logging.Error("没有可用的规则，请检查过滤条件")
		os.Exit(1)
	}

	// 打印规则信息
	config.PrintRulesInfo(filteredRules)

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
		outputProcessor := utils.NewOutputProcessor(cmdConfig)
		if err := outputProcessor.ProcessResults(results); err != nil {
			logging.Errorf("输出结果失败: %v", err)
			os.Exit(1)
		}
	} else {
		logging.Info("未发现任何敏感信息")
	}

	logging.Info("程序执行完成")
}
