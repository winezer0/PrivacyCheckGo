package baserule

import (
	"fmt"
	"privacycheck/logging"
	"regexp"
	"strings"
)

// ValidateRules 验证规则配置
func (config *RulesConfig) ValidateRules() error {
	logging.Info("开始验证规则...")

	var invalidRules []string
	validRulesCount := 0

	for _, group := range config.Rules {
		for _, rule := range group.Rule {
			// 检查loaded字段(默认为True)
			if !rule.Loaded {
				continue
			}

			validRulesCount++

			// 检查必要字段
			if rule.Name == "" {
				invalidRules = append(invalidRules, fmt.Sprintf("规则组 %s: 规则缺少name字段", group.Group))
				continue
			}

			if rule.FRegex == "" {
				invalidRules = append(invalidRules, fmt.Sprintf("规则组 %s, 规则 %s: 缺少f_regex字段", group.Group, rule.Name))
				continue
			}

			// 验证正则表达式
			if _, err := regexp.Compile(rule.FRegex); err != nil {
				invalidRules = append(invalidRules, fmt.Sprintf("规则组 %s, 规则 %s: 正则表达式无效 - %v", group.Group, rule.Name, err))
			}
		}
	}

	if len(invalidRules) > 0 {
		logging.Error("发现无效的规则:")
		for _, invalid := range invalidRules {
			logging.Error(invalid)
		}
		return fmt.Errorf("发现 %d 个无效规则", len(invalidRules))
	}

	logging.Infof("规则验证通过！共 %d 个有效规则", validRulesCount)
	return nil
}

// FilterRules 过滤规则
func FilterRules(config *RulesConfig, filterGroups, filterNames []string, sensitiveOnly bool) RuleMap {
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

	for _, group := range config.Rules {
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

			// 设置默认的上下文长度
			if rule.ContextLeft == 0 && rule.ContextRight == 0 && rule.Sensitive {
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

// PrintRulesInfo 打印规则信息
func PrintRulesInfo(rules RuleMap) {
	logging.Info("本次扫描使用的规则:")

	for groupName, ruleList := range rules {
		for _, rule := range ruleList {
			regex := rule.FRegex
			if len(regex) > 50 {
				regex = regex[:47] + "..."
			}
			logging.Infof("%s: %s: %s", groupName, rule.Name, regex)
		}
	}

	logging.Info(strings.Repeat("=", 50))
}

// CountRules 计算规则数量
func CountRules(rules RuleMap) int {
	count := 0
	for _, ruleList := range rules {
		count += len(ruleList)
	}
	return count
}
