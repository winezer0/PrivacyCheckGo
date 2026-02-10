package baserule

import (
	"fmt"
	"strings"

	"github.com/winezer0/xutils/logging"
)

// ValidateRules 验证规则配置
func (c *RuleConfig) ValidateRules() error {
	logging.Debug("Start verifying the rules...")

	var invalidRules []string
	validRulesCount := 0

	for _, group := range c.Rules {
		for _, rule := range group.Rule {
			// 检查loaded字段(默认为True)
			if !rule.Loaded {
				continue
			}

			validRulesCount++

			// 检查必要字段
			if rule.Name == "" {
				invalidRules = append(invalidRules, fmt.Sprintf("Rule Group %s: The rule lacks the [name] field", group.Group))
				continue
			}

			if rule.FRegex == "" {
				invalidRules = append(invalidRules, fmt.Sprintf("Rule Group %s, Rule %s: The rule lacks the [f_regex] field.", group.Group, rule.Name))
				continue
			}

			// 验证正则表达式，使用回退机制
			if _, err := TryCompileWithFallback(rule.FRegex); err != nil {
				invalidRules = append(invalidRules, fmt.Sprintf("Rule Group %s, Rule %s: The rule f_regex [f_regex] compile is error:%v", group.Group, rule.Name, err))
			}
		}
	}

	if len(invalidRules) > 0 {
		logging.Error("Discovering invalid rules:")
		for _, invalid := range invalidRules {
			logging.Error(invalid)
		}
		return fmt.Errorf("found invalid rules: %d", len(invalidRules))
	}

	logging.Infof("rules verify pass! total valid rules: %d", validRulesCount)
	return nil
}

// FilterRules 过滤规则
func (c *RuleConfig) FilterRules(filterGroups, filterNames []string, sensitiveOnly bool) RuleMap {
	result := make(RuleMap)

	// 转换过滤条件为小写
	var lowerFilterGroups, lowerFilterNames []string
	for _, group := range filterGroups {
		if strings.TrimSpace(group) != "" {
			lowerFilterGroups = append(lowerFilterGroups, strings.ToLower(strings.TrimSpace(group)))
		}
	}
	for _, name := range filterNames {
		if strings.TrimSpace(name) != "" {
			lowerFilterNames = append(lowerFilterNames, strings.ToLower(strings.TrimSpace(name)))
		}
	}

	for _, group := range c.Rules {
		// 按照group_name进行过滤
		if len(lowerFilterGroups) > 0 {
			groupMatched := false
			for _, filterGroup := range lowerFilterGroups {
				if strings.Contains(strings.ToLower(group.Group), filterGroup) {
					groupMatched = true
					break
				}
			}
			if !groupMatched {
				continue
			}
		}

		var filteredRules []Rule
		for _, rule := range group.Rule {
			// 排除空规则或未加载的规则
			if rule.Name == "" || rule.FRegex == "" || !rule.Loaded {
				continue
			}

			// 仅敏感模式下排除非敏感信息的规则
			if sensitiveOnly && !rule.Sensitive {
				continue
			}

			// 按名称关键字过滤
			if len(lowerFilterNames) > 0 {
				nameMatched := false
				for _, filterName := range lowerFilterNames {
					if strings.Contains(strings.ToLower(rule.Name), filterName) {
						nameMatched = true
						break
					}
				}
				if !nameMatched {
					continue
				}
			}

			// 对于敏感信息规则设置默认的上下文长度
			if rule.Sensitive && rule.ContextLeft <= 0 && rule.ContextRight <= 0 {
				rule.ContextLeft = 50
				rule.ContextRight = 50
			}

			filteredRules = append(filteredRules, rule)
		}

		if len(filteredRules) > 0 {
			result[group.Group] = filteredRules
		}
	}

	return result
}
