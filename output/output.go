package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"privacycheck/config"
	"privacycheck/core"
	"privacycheck/pkg/logging"
	"reflect"
	"strings"
)

// Output 输出处理器
type Output struct {
	config *config.Config
}

// NewOutput 创建输出处理器
func NewOutput(config *config.Config) *Output {
	return &Output{config: config}
}

// ProcessResults 处理扫描结果
func (p *Output) ProcessResults(results []core.ScanResult, stats *core.ScanStats) error {
	if len(results) == 0 {
		logging.Info("没有发现任何结果")
		return nil
	}

	// 输出统计信息
	p.printStatistics(results)

	// 格式化结果
	if p.config.FormatResults {
		results = p.formatResults(results)
	}

	// 过滤黑名单匹配
	if len(p.config.BlockMatches) > 0 {
		results = p.filterBlockMatches(results)
	}

	// 按组分组输出
	var groupedResults map[string][]core.ScanResult
	if p.config.OutputGroup {
		groupedResults = p.groupByField(results, "group")
	} else {
		groupedResults = map[string][]core.ScanResult{"": results}
	}

	// 过滤输出字段
	if len(p.config.OutputKeys) > 0 {
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

// formatResults 格式化结果
func (p *Output) formatResults(results []core.ScanResult) []core.ScanResult {
	for i := range results {
		results[i].Match = p.stripString(results[i].Match)
		results[i].Context = p.stripString(results[i].Context)
		results[i].File = p.stripString(results[i].File)
		results[i].Group = p.stripString(results[i].Group)
		results[i].RuleName = p.stripString(results[i].RuleName)
	}
	return results
}

// stripString 清理字符串
func (p *Output) stripString(s string) string {
	// 去除首尾的引号、括号、空格等
	s = strings.Trim(s, "'\"()[]{}  \n\r\t")
	// 替换转义的斜杠
	s = strings.ReplaceAll(s, `\/`, "/")
	return s
}

// filterBlockMatches 过滤黑名单匹配
func (p *Output) filterBlockMatches(results []core.ScanResult) []core.ScanResult {
	var filtered []core.ScanResult

	for _, result := range results {
		blocked := false
		for _, blockWord := range p.config.BlockMatches {
			if strings.Contains(result.Match, blockWord) {
				blocked = true
				break
			}
		}
		if !blocked {
			filtered = append(filtered, result)
		}
	}

	logging.Infof("黑名单过滤: %d -> %d", len(results), len(filtered))
	return filtered
}

// groupByField 按字段分组
func (p *Output) groupByField(results []core.ScanResult, field string) map[string][]core.ScanResult {
	groups := make(map[string][]core.ScanResult)

	for _, result := range results {
		var key string
		switch field {
		case "group":
			key = result.Group
		case "rule_name":
			key = result.RuleName
		case "file":
			key = result.File
		default:
			key = ""
		}

		groups[key] = append(groups[key], result)
	}

	return groups
}

// filterOutputKeys 过滤输出字段
func (p *Output) filterOutputKeys(results []core.ScanResult) []core.ScanResult {
	if len(p.config.OutputKeys) == 0 {
		return results
	}

	// 创建字段映射
	keyMap := make(map[string]bool)
	for _, key := range p.config.OutputKeys {
		keyMap[key] = true
	}

	// 过滤字段
	var filtered []core.ScanResult
	for _, result := range results {
		newResult := core.ScanResult{}

		if keyMap["file"] {
			newResult.File = result.File
		}
		if keyMap["group"] {
			newResult.Group = result.Group
		}
		if keyMap["rule_name"] {
			newResult.RuleName = result.RuleName
		}
		if keyMap["match"] {
			newResult.Match = result.Match
		}
		if keyMap["context"] {
			newResult.Context = result.Context
		}
		if keyMap["position"] {
			newResult.Position = result.Position
		}
		if keyMap["line_number"] {
			newResult.LineNumber = result.LineNumber
		}
		if keyMap["sensitive"] {
			newResult.Sensitive = result.Sensitive
		}

		filtered = append(filtered, newResult)
	}

	return filtered
}

// outputGroup 输出单个组的结果
func (p *Output) outputGroup(groupName string, results []core.ScanResult) error {
	// 生成输出文件名
	baseOutput := p.config.OutputFile
	if baseOutput == "" {
		baseOutput = p.config.ProjectName
	}

	// 移除现有扩展名
	if ext := filepath.Ext(baseOutput); ext != "" {
		baseOutput = strings.TrimSuffix(baseOutput, ext)
	}

	var outputFile string
	if groupName != "" {
		outputFile = fmt.Sprintf("%s.%s.%s", baseOutput, groupName, p.config.OutputFormat)
	} else {
		outputFile = fmt.Sprintf("%s.%s", baseOutput, p.config.OutputFormat)
	}

	// 确保输出目录存在
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 根据格式输出
	switch p.config.OutputFormat {
	case "csv":
		if err := p.writeCSV(outputFile, results); err != nil {
			return err
		}
	case "json":
		if err := p.writeJSON(outputFile, results); err != nil {
			return err
		}
	default:
		return fmt.Errorf("不支持的输出格式: %s", p.config.OutputFormat)
	}

	logging.Infof("分析结果 [group:%s|format:%s] 已保存至: %s",
		groupName, p.config.OutputFormat, outputFile)

	return nil
}

// writeJSON 写入JSON文件
func (p *Output) writeJSON(filename string, results []core.ScanResult) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建JSON文件失败: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(results); err != nil {
		return fmt.Errorf("写入JSON数据失败: %w", err)
	}

	return nil
}

// writeCSV 写入CSV文件
func (p *Output) writeCSV(filename string, results []core.ScanResult) error {
	if len(results) == 0 {
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建CSV文件失败: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := p.getCSVHeaders(results[0])
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("写入CSV表头失败: %w", err)
	}

	// 写入数据
	for _, result := range results {
		record := p.resultToCSVRecord(result, headers)
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("写入CSV数据失败: %w", err)
		}
	}

	return nil
}

// getCSVHeaders 获取CSV表头
func (p *Output) getCSVHeaders(result core.ScanResult) []string {
	var headers []string

	// 如果指定了输出字段，使用指定的字段顺序
	if len(p.config.OutputKeys) > 0 {
		return p.config.OutputKeys
	}

	// 否则使用默认字段顺序
	v := reflect.ValueOf(result)
	t := reflect.TypeOf(result)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			headers = append(headers, jsonTag)
		}
	}

	return headers
}

// resultToCSVRecord 将结果转换为CSV记录
func (p *Output) resultToCSVRecord(result core.ScanResult, headers []string) []string {
	record := make([]string, len(headers))

	for i, header := range headers {
		switch header {
		case "file":
			record[i] = result.File
		case "group":
			record[i] = result.Group
		case "rule_name":
			record[i] = result.RuleName
		case "match":
			record[i] = result.Match
		case "context":
			record[i] = result.Context
		case "position":
			record[i] = fmt.Sprintf("%d", result.Position)
		case "line_number":
			record[i] = fmt.Sprintf("%d", result.LineNumber)
		case "sensitive":
			record[i] = fmt.Sprintf("%t", result.Sensitive)
		default:
			record[i] = ""
		}
	}

	return record
}

// printStatistics 打印统计信息
func (p *Output) printStatistics(results []core.ScanResult) {
	// 统计各种信息
	sensitiveCount := 0
	fileCount := make(map[string]bool)
	groupCount := make(map[string]int)
	ruleCount := make(map[string]int)

	for _, result := range results {
		// 文件统计
		fileCount[result.File] = true

		// 敏感信息统计
		if result.Sensitive {
			sensitiveCount++
		}

		// 规则组统计
		groupCount[result.Group]++

		// 规则统计
		ruleCount[result.RuleName]++
	}

	logging.Info("=== 扫描结果统计 ===")
	logging.Infof("总结果数: %d", len(results))
	logging.Infof("敏感信息数: %d", sensitiveCount)
	logging.Infof("涉及文件数: %d", len(fileCount))
	logging.Infof("规则组数: %d", len(groupCount))

	// 按规则组统计
	logging.Info("按规则组统计:")
	for group, count := range groupCount {
		logging.Infof("  %s: %d", group, count)
	}

	// 显示前5个最常触发的规则
	logging.Info("最常触发的规则:")
	type ruleStats struct {
		name  string
		count int
	}

	var topRules []ruleStats
	for rule, count := range ruleCount {
		topRules = append(topRules, ruleStats{rule, count})
	}

	// 简单排序（冒泡排序）
	for i := 0; i < len(topRules)-1; i++ {
		for j := 0; j < len(topRules)-1-i; j++ {
			if topRules[j].count < topRules[j+1].count {
				topRules[j], topRules[j+1] = topRules[j+1], topRules[j]
			}
		}
	}

	// 显示前5个
	maxShow := 5
	if len(topRules) < maxShow {
		maxShow = len(topRules)
	}

	for i := 0; i < maxShow; i++ {
		logging.Infof("  %s: %d", topRules[i].name, topRules[i].count)
	}

	logging.Info("==================")
}
