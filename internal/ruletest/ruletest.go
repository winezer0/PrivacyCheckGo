package ruletest

import (
	"fmt"
	"github.com/winezer0/xutils/logging"
	"os"
	"path/filepath"
	"privacycheck/internal/baserule"
	"strings"
)

// RunRuleTest 测试所有规则并生成测试报告
func RunRuleTest(rulesFile string, rules []baserule.Rules) {
	logging.Info("Running rule test mode...")

	// 收集测试结果
	var (
		compileErrorRules []string
		noSampleCodeRules []string
		validRules        []string
		totalRules        int
	)

	// 遍历所有规则组和规则
	for _, group := range rules {
		for _, rule := range group.Rule {
			// 跳过未加载的规则
			if !rule.Loaded {
				continue
			}

			totalRules++
			ruleIdentifier := fmt.Sprintf("%s: %s", group.Group, rule.Name)

			// 测试正则表达式编译
			_, err := baserule.TryCompileWithFallback(rule.FRegex)
			if err != nil {
				compileErrorRules = append(compileErrorRules, fmt.Sprintf("%s - 编译错误: %v", ruleIdentifier, err))
				continue
			}

			// 检查是否有 SampleCode
			if rule.SampleCode == "" {
				noSampleCodeRules = append(noSampleCodeRules, ruleIdentifier)
				continue
			}

			// 测试正则表达式是否匹配 SampleCode
			matcher, _ := baserule.TryCompileWithFallback(rule.FRegex)
			match, err := matcher.MatchString(rule.SampleCode)
			if err != nil || !match {
				compileErrorRules = append(compileErrorRules, fmt.Sprintf("%s - SampleCode 匹配失败: %v", ruleIdentifier, err))
				continue
			}

			// 规则有效
			validRules = append(validRules, ruleIdentifier)
		}
	}

	// 生成测试报告
	reportFile := fmt.Sprintf("%s_test.md", strings.TrimSuffix(rulesFile, filepath.Ext(rulesFile)))
	reportContent := genTestReport(rulesFile, totalRules, compileErrorRules, noSampleCodeRules, validRules)

	// 保存报告文件
	if err := os.WriteFile(reportFile, []byte(reportContent), 0644); err != nil {
		logging.Fatalf("Failed to save test report: %v", err)
	}

	// 输出测试结果摘要
	logging.Infof("Rule test completed!")
	logging.Infof("Total rules tested: %d", totalRules)
	logging.Infof("Compile error rules: %d", len(compileErrorRules))
	logging.Infof("No SampleCode rules: %d", len(noSampleCodeRules))
	logging.Infof("Valid rules: %d", len(validRules))
	logging.Infof("Test report saved to: %s", reportFile)
}
