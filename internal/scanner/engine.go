package scanner

import (
	"fmt"
	"github.com/winezer0/xutils/logging"
	"strings"

	"privacycheck/internal/baserule"
)

// RuleEngine 规则引擎
type RuleEngine struct {
	rules       baserule.RuleMap
	compiledReg map[string]baserule.RegexMatcher
}

// NewRuleEngine 创建新的规则引擎
func NewRuleEngine(rules baserule.RuleMap) (*RuleEngine, error) {
	engine := &RuleEngine{
		rules:       rules,
		compiledReg: make(map[string]baserule.RegexMatcher),
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
			//IgnoreCase
			pattern = "(?i)" + pattern
			// 添加多行模式
			pattern = "(?m)" + pattern

			var matcher baserule.RegexMatcher
			var err error

			// 根据规则指定的引擎
			if rule.Engine == "dfa" || rule.Engine == "nfa" {
				rule.Engine = string(baserule.RegexEngineJava)
			}

			if rule.Engine != "" {
				// 使用指定的引擎
				engine := baserule.RegexEngine(rule.Engine)
				matcher, err = baserule.NewRegexMatcher(pattern, engine)
				if err != nil {
					return fmt.Errorf("failed to compile regex with specified engine [%s:%s]: %w", groupName, rule.Name, err)
				}
			} else {
				// 引擎为空，先尝试 Go 引擎，失败则使用 Java 引擎
				matcher, err = baserule.TryCompileWithFallback(pattern)
				if err != nil {
					return fmt.Errorf("failed to compile regex with fallback [%s:%s]: %w", groupName, rule.Name, err)
				}
			}

			e.compiledReg[key] = matcher
		}
	}

	logging.Infof("successfully compiled %d regex patterns", len(e.compiledReg))
	return nil
}

// ApplyRules 对内容应用所有规则，支持指定偏移量和起始行号
func (e *RuleEngine) ApplyRules(content, filePath string, positionOffset int, startLineNumber int) []ScanResult {
	var results []ScanResult

	for groupName, ruleList := range e.rules {
		for i, rule := range ruleList {
			key := fmt.Sprintf("%s_%d", groupName, i)
			regex := e.compiledReg[key]

			ruleResults := e.applyRule(rule, regex, content, groupName, filePath, positionOffset, startLineNumber)
			results = append(results, ruleResults...)
		}
	}

	return results
}

// applyRule 应用单个规则，支持指定偏移量和起始行号
func (e *RuleEngine) applyRule(rule baserule.Rule, matcher baserule.RegexMatcher, content, groupName, filePath string, positionOffset int, startLineNumber int) []ScanResult {
	var results []ScanResult

	// 查找第一个匹配
	match, err := matcher.FindStringMatch(content)
	if err != nil {
		logging.Warnf("error matching regex [%s:%s]: %v", groupName, rule.Name, err)
		return results
	}

	// 遍历所有匹配
	for match != nil {
		matchedText := match.String()

		// 过滤过短的匹配
		if len(strings.TrimSpace(matchedText)) <= 5 {
			nextMatch, err := match.FindNextMatch()
			if err != nil {
				break
			}
			match = nextMatch
			continue
		}

		// 查找匹配在原文中的位置
		start := strings.Index(content, matchedText)
		if start == -1 {
			nextMatch, err := match.FindNextMatch()
			if err != nil {
				break
			}
			match = nextMatch
			continue
		}

		end := start + len(matchedText)

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

		result := ScanResult{
			File:       filePath,
			Group:      groupName,
			RuleName:   rule.Name,
			Match:      matchedText,
			Context:    context,
			Position:   positionOffset + start,                                 // 加上位置偏移
			LineNumber: startLineNumber + strings.Count(content[:start], "\n"), // 计算行号（考虑起始行号偏移）
			Sensitive:  rule.Sensitive,
		}

		results = append(results, result)

		// 查找下一个匹配
		nextMatch, err := match.FindNextMatch()
		if err != nil {
			break
		}
		match = nextMatch
	}

	return results
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
