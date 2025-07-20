package scanner

import (
	"fmt"
	"regexp"
	"strings"

	"privacycheck/core"
	"privacycheck/logging"
)

// RuleEngine 规则引擎
type RuleEngine struct {
	rules       map[string][]core.Rule
	compiledReg map[string]*regexp.Regexp
}

// NewRuleEngine 创建新的规则引擎
func NewRuleEngine(rules map[string][]core.Rule) (*RuleEngine, error) {
	engine := &RuleEngine{
		rules:       rules,
		compiledReg: make(map[string]*regexp.Regexp),
	}

	// 预编译所有正则表达式
	if err := engine.compileRegexes(); err != nil {
		return nil, err
	}

	return engine, nil
}

// compileRegexes 预编译所有正则表达式
func (e *RuleEngine) compileRegexes() error {
	for groupName, ruleList := range e.rules {
		for i, rule := range ruleList {
			key := fmt.Sprintf("%s_%d", groupName, i)

			// 编译正则表达式
			pattern := rule.FRegex
			if rule.IgnoreCase {
				pattern = "(?i)" + pattern
			}
			// 添加多行模式
			pattern = "(?m)" + pattern

			regex, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("编译正则表达式失败 [%s:%s]: %w", groupName, rule.Name, err)
			}

			e.compiledReg[key] = regex
		}
	}

	logging.Infof("成功编译 %d 个正则表达式", len(e.compiledReg))
	return nil
}

// ApplyRules 对内容应用所有规则
func (e *RuleEngine) ApplyRules(content, filePath string) []core.ScanResult {
	var results []core.ScanResult

	for groupName, ruleList := range e.rules {
		for i, rule := range ruleList {
			key := fmt.Sprintf("%s_%d", groupName, i)
			regex := e.compiledReg[key]

			ruleResults := e.applyRule(rule, regex, content, groupName, filePath)
			results = append(results, ruleResults...)
		}
	}

	return results
}

// applyRule 应用单个规则
func (e *RuleEngine) applyRule(rule core.Rule, regex *regexp.Regexp, content, groupName, filePath string) []core.ScanResult {
	var results []core.ScanResult

	// 查找所有匹配
	matches := regex.FindAllStringSubmatchIndex(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		start, end := match[0], match[1]
		matchedText := content[start:end]

		// 过滤过短的匹配
		if len(strings.TrimSpace(matchedText)) <= 5 {
			continue
		}

		// 计算上下文
		contextLeft := rule.ContextLeft
		contextRight := rule.ContextRight

		// 如果是敏感信息且未设置上下文，使用默认值
		if rule.Sensitive && contextLeft == 0 && contextRight == 0 {
			contextLeft = 50
			contextRight = 50
		}

		contextStart := max(0, start-contextLeft)
		contextEnd := min(len(content), end+contextRight)
		context := content[contextStart:contextEnd]

		// 计算行号
		lineNumber := strings.Count(content[:start], "\n") + 1

		result := core.ScanResult{
			File:       filePath,
			Group:      groupName,
			RuleName:   rule.Name,
			Match:      matchedText,
			Context:    context,
			Position:   start,
			LineNumber: lineNumber,
			Sensitive:  rule.Sensitive,
		}

		results = append(results, result)
	}

	return results
}

// ApplyRuleToChunk 对数据块应用规则（用于chunk模式）
func (e *RuleEngine) ApplyRuleToChunk(content, filePath string, chunkOffset int) []core.ScanResult {
	var results []core.ScanResult

	for groupName, ruleList := range e.rules {
		for i, rule := range ruleList {
			key := fmt.Sprintf("%s_%d", groupName, i)
			regex := e.compiledReg[key]

			chunkResults := e.applyRuleToChunk(rule, regex, content, groupName, filePath, chunkOffset)
			results = append(results, chunkResults...)
		}
	}

	return results
}

// applyRuleToChunk 对数据块应用单个规则
func (e *RuleEngine) applyRuleToChunk(rule core.Rule, regex *regexp.Regexp, content, groupName, filePath string, chunkOffset int) []core.ScanResult {
	var results []core.ScanResult

	matches := regex.FindAllStringSubmatchIndex(content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		start, end := match[0], match[1]
		matchedText := content[start:end]

		if len(strings.TrimSpace(matchedText)) <= 5 {
			continue
		}

		// 计算上下文（在chunk内）
		contextLeft := rule.ContextLeft
		contextRight := rule.ContextRight

		if rule.Sensitive && contextLeft == 0 && contextRight == 0 {
			contextLeft = 50
			contextRight = 50
		}

		contextStart := max(0, start-contextLeft)
		contextEnd := min(len(content), end+contextRight)
		context := content[contextStart:contextEnd]

		// 计算在整个文件中的行号（近似）
		lineNumber := strings.Count(content[:start], "\n") + 1

		result := core.ScanResult{
			File:       filePath,
			Group:      groupName,
			RuleName:   rule.Name,
			Match:      matchedText,
			Context:    context,
			Position:   chunkOffset + start, // 加上chunk偏移量
			LineNumber: lineNumber,          // 注意：这是chunk内的行号，不是文件的绝对行号
			Sensitive:  rule.Sensitive,
		}

		results = append(results, result)
	}

	return results
}

// GetRulesCount 获取规则总数
func (e *RuleEngine) GetRulesCount() int {
	count := 0
	for _, ruleList := range e.rules {
		count += len(ruleList)
	}
	return count
}

// GetGroupsCount 获取规则组数量
func (e *RuleEngine) GetGroupsCount() int {
	return len(e.rules)
}

// ValidateRule 验证单个规则
func ValidateRule(rule core.Rule) error {
	if rule.Name == "" {
		return fmt.Errorf("规则名称不能为空")
	}

	if rule.FRegex == "" {
		return fmt.Errorf("规则 %s: 正则表达式不能为空", rule.Name)
	}

	// 验证正则表达式语法
	pattern := rule.FRegex
	if rule.IgnoreCase {
		pattern = "(?i)" + pattern
	}

	if _, err := regexp.Compile(pattern); err != nil {
		return fmt.Errorf("规则 %s: 正则表达式语法错误 - %w", rule.Name, err)
	}

	return nil
}

// 辅助函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
