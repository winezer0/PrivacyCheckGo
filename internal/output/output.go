package output

import (
	"fmt"
	"path/filepath"
	"privacycheck/internal/scanner"
	"privacycheck/pkg/fileutils"
	"privacycheck/pkg/logging"
	"strings"
)

// Output 输出处理器配置
type Output struct {
	OutputFile    string
	OutputGroup   bool
	OutputKeys    []string
	OutputFormat  string
	FormatResults bool
	BlockMatches  []string
	ProjectName   string
}

// ProcessResults 处理扫描结果
func (p *Output) ProcessResults(results []scanner.ScanResult, stats *scanner.ScanStats) error {
	if len(results) == 0 {
		logging.Info("没有发现任何结果")
		return nil
	}

	// 输出统计信息
	p.printStatistics(results)

	// 格式化结果
	if p.FormatResults {
		results = p.formatResults(results)
	}

	// 过滤黑名单匹配
	if len(p.BlockMatches) > 0 {
		results = p.filterBlockMatches(results)
	}

	// 按组分组输出
	var groupedResults map[string][]scanner.ScanResult
	if p.OutputGroup {
		groupedResults = p.groupByField(results, "group")
	} else {
		groupedResults = map[string][]scanner.ScanResult{"": results}
	}

	// 过滤输出字段
	if len(p.OutputKeys) > 0 {
		for groupName, groupResults := range groupedResults {
			groupedResults[groupName] = p.filterOutputKeys(groupResults)
		}
	}

	// 输出结果
	for groupName, groupResults := range groupedResults {
		if err := p.outputGroup(groupName, groupResults); err != nil {
			return fmt.Errorf("输出结果失败: %w", err)
		}
	}

	return nil
}

// outputGroup 输出单个组的结果
func (p *Output) outputGroup(groupName string, results []scanner.ScanResult) error {
	// 生成输出文件名
	baseOutput := p.OutputFile
	if baseOutput == "" {
		baseOutput = p.ProjectName
	}

	// 移除现有扩展名
	if ext := filepath.Ext(baseOutput); ext != "" {
		baseOutput = strings.TrimSuffix(baseOutput, ext)
	}

	var outputFile string
	if groupName != "" {
		outputFile = fmt.Sprintf("%s.%s.%s", baseOutput, groupName, p.OutputFormat)
	} else {
		outputFile = fmt.Sprintf("%s.%s", baseOutput, p.OutputFormat)
	}

	// 确保输出目录存在
	if err := fileutils.EnsureDir(filepath.Dir(outputFile)); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 根据格式输出
	switch p.OutputFormat {
	case "csv":
		if err := p.writeCSV(outputFile, results); err != nil {
			return err
		}
	case "json":
		if err := p.writeJSON(outputFile, results); err != nil {
			return err
		}
	default:
		return fmt.Errorf("不支持的输出格式: %s", p.OutputFormat)
	}

	logging.Infof("分析结果 [group:%s|format:%s] 已保存至: %s",
		groupName, p.OutputFormat, outputFile)

	return nil
}
