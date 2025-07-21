package baserule

import (
	"privacycheck/pkg/logging"
	"strings"
)

type RuleMap map[string][]Rule

// PrintRulesInfo 打印规则信息
func (m *RuleMap) PrintRulesInfo() {
	logging.Info("The rules used for this scan:")

	for groupName, ruleList := range *m {
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
func (m *RuleMap) CountRules() int {
	count := 0
	for _, ruleList := range *m {
		count += len(ruleList)
	}
	return count
}
