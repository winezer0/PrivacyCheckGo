package output

import (
	"privacycheck/internal/scanner"
	"privacycheck/pkg/logging"
	"strings"
)

// filterBlockMatches 过滤黑名单匹配
func (p *Output) filterBlockMatches(results []scanner.ScanResult) []scanner.ScanResult {
	var filtered []scanner.ScanResult

	for _, result := range results {
		blocked := false
		for _, blockWord := range p.BlockMatches {
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
func (p *Output) groupByField(results []scanner.ScanResult, field string) map[string][]scanner.ScanResult {
	groups := make(map[string][]scanner.ScanResult)

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
func (p *Output) filterOutputKeys(results []scanner.ScanResult) []scanner.ScanResult {
	if len(p.OutputKeys) == 0 {
		return results
	}

	// 创建字段映射
	keyMap := make(map[string]bool)
	for _, key := range p.OutputKeys {
		keyMap[key] = true
	}

	// 过滤字段
	var filtered []scanner.ScanResult
	for _, result := range results {
		newResult := p.filterSingleResult(result, keyMap)
		filtered = append(filtered, newResult)
	}

	return filtered
}

// filterSingleResult 过滤单个结果的字段
func (p *Output) filterSingleResult(result scanner.ScanResult, keyMap map[string]bool) scanner.ScanResult {
	newResult := scanner.ScanResult{}

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

	return newResult
}
