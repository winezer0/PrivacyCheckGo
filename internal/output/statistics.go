package output

import (
	"privacycheck/internal/scanner"
	"privacycheck/pkg/logging"
)

// ruleStats 规则统计结构
type ruleStats struct {
	name  string
	count int
}

// printStatistics 打印统计信息
func (p *Output) printStatistics(results []scanner.ScanResult) {
	stats := p.calculateStatistics(results)
	p.displayStatistics(stats, results)
}

// statisticsData 统计数据结构
type statisticsData struct {
	sensitiveCount int
	fileCount      map[string]bool
	groupCount     map[string]int
	ruleCount      map[string]int
}

// calculateStatistics 计算统计信息
func (p *Output) calculateStatistics(results []scanner.ScanResult) *statisticsData {
	stats := &statisticsData{
		fileCount:  make(map[string]bool),
		groupCount: make(map[string]int),
		ruleCount:  make(map[string]int),
	}

	for _, result := range results {
		// 文件统计
		stats.fileCount[result.File] = true

		// 敏感信息统计
		if result.Sensitive {
			stats.sensitiveCount++
		}

		// 规则组统计
		stats.groupCount[result.Group]++

		// 规则统计
		stats.ruleCount[result.RuleName]++
	}

	return stats
}

// displayStatistics 显示统计信息
func (p *Output) displayStatistics(stats *statisticsData, results []scanner.ScanResult) {
	logging.Info("=== Scan Results Statistics ===")
	logging.Infof("total results: %d", len(results))
	logging.Infof("sensitive information: %d", stats.sensitiveCount)
	logging.Infof("files involved: %d", len(stats.fileCount))
	logging.Infof("rule groups: %d", len(stats.groupCount))

	// 按规则组统计
	p.displayGroupStatistics(stats.groupCount)

	// 显示最常触发的规则
	p.displayTopRules(stats.ruleCount)

	logging.Info("==================")
}

// displayGroupStatistics 显示规则组统计
func (p *Output) displayGroupStatistics(groupCount map[string]int) {
	logging.Info("statistics by rule group:")
	for group, count := range groupCount {
		logging.Infof("  %s: %d", group, count)
	}
}

// displayTopRules 显示最常触发的规则
func (p *Output) displayTopRules(ruleCount map[string]int) {
	logging.Info("most frequently triggered rules:")

	topRules := p.sortRulesByCount(ruleCount)
	p.showTopRules(topRules, 5)
}

// sortRulesByCount 按计数排序规则
func (p *Output) sortRulesByCount(ruleCount map[string]int) []ruleStats {
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

	return topRules
}

// showTopRules 显示前N个规则
func (p *Output) showTopRules(topRules []ruleStats, maxShow int) {
	if len(topRules) < maxShow {
		maxShow = len(topRules)
	}

	for i := 0; i < maxShow; i++ {
		logging.Infof("  %s: %d", topRules[i].name, topRules[i].count)
	}
}
