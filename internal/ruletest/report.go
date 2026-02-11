package ruletest

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// genTestReport 生成规则测试报告
func genTestReport(rulesFile string, totalRules int, compileErrorRules []string, noSampleCodeRules []string, validRules []string) string {
	var buf strings.Builder

	// 报告标题
	buf.WriteString(fmt.Sprintf("# 规则测试报告\n\n"))
	buf.WriteString(fmt.Sprintf("## 测试配置\n\n"))
	buf.WriteString(fmt.Sprintf("- **规则文件**: %s\n", rulesFile))
	buf.WriteString(fmt.Sprintf("- **测试日期**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("- **总规则数**: %d\n", totalRules))
	buf.WriteString(fmt.Sprintf("- **Go 版本**: %s\n", runtime.Version()))

	// 测试结果摘要
	buf.WriteString("## 测试结果摘要\n\n")
	buf.WriteString(fmt.Sprintf("| 状态 | 数量 | 百分比 | 状态码 |\n"))
	buf.WriteString(fmt.Sprintf("|------|------|--------|--------|\n"))
	buf.WriteString(fmt.Sprintf("| 有效规则 | %d | %.2f%% | ✅ |\n", len(validRules), float64(len(validRules))/float64(totalRules)*100))
	buf.WriteString(fmt.Sprintf("| 无 SampleCode | %d | %.2f%% | ⚠️ |\n", len(noSampleCodeRules), float64(len(noSampleCodeRules))/float64(totalRules)*100))
	buf.WriteString(fmt.Sprintf("| 编译错误 | %d | %.2f%% | ❌ |\n\n", len(compileErrorRules), float64(len(compileErrorRules))/float64(totalRules)*100))

	// 测试结果分析
	buf.WriteString("## 测试结果分析\n\n")
	if len(validRules) == totalRules {
		buf.WriteString("### 🎉 测试通过！\n\n")
		buf.WriteString("所有规则都通过了测试，没有发现任何问题。\n\n")
	} else if len(validRules) > len(compileErrorRules)+len(noSampleCodeRules) {
		buf.WriteString("### 📊 测试基本通过\n\n")
		buf.WriteString("大部分规则通过了测试，但仍有一些问题需要解决。\n\n")
	} else {
		buf.WriteString("### ⚠️ 测试未通过\n\n")
		buf.WriteString("有较多规则未通过测试，需要仔细检查和修复。\n\n")
	}

	// 编译错误规则
	buf.WriteString("## 编译错误规则\n\n")
	if len(compileErrorRules) > 0 {
		buf.WriteString(fmt.Sprintf("发现 %d 个规则编译错误:\n\n", len(compileErrorRules)))
		buf.WriteString("| 规则 | 错误信息 |\n")
		buf.WriteString("|------|----------|\n")
		for _, rule := range compileErrorRules {
			buf.WriteString(fmt.Sprintf("| %s | |\n", rule))
		}
		buf.WriteString("\n")
		buf.WriteString("### 修复建议\n\n")
		buf.WriteString("1. 检查正则表达式语法是否正确\n")
		buf.WriteString("2. 确保 SampleCode 能够被正则表达式匹配\n")
		buf.WriteString("3. 验证正则表达式是否符合 Go 或 Java 正则语法规范\n\n")
	} else {
		buf.WriteString("未发现编译错误规则。\n\n")
	}

	// 无 SampleCode 规则
	buf.WriteString("## 无 SampleCode 规则\n\n")
	if len(noSampleCodeRules) > 0 {
		buf.WriteString(fmt.Sprintf("发现 %d 个规则缺少 SampleCode:\n\n", len(noSampleCodeRules)))
		buf.WriteString("| 规则 |\n")
		buf.WriteString("|------|\n")
		for _, rule := range noSampleCodeRules {
			buf.WriteString(fmt.Sprintf("| %s |\n", rule))
		}
		buf.WriteString("\n")
		buf.WriteString("### 修复建议\n\n")
		buf.WriteString("1. 为每个规则添加 sample_code 字段\n")
		buf.WriteString("2. 确保 SampleCode 能够代表该规则要匹配的数据格式\n")
		buf.WriteString("3. 验证 SampleCode 能够被正则表达式匹配\n\n")
	} else {
		buf.WriteString("所有规则都有 SampleCode。\n\n")
	}

	// 有效规则
	buf.WriteString("## 有效规则\n\n")
	if len(validRules) > 0 {
		buf.WriteString(fmt.Sprintf("发现 %d 个有效规则:\n\n", len(validRules)))
		buf.WriteString("| 规则 |\n")
		buf.WriteString("|------|\n")
		for _, rule := range validRules {
			buf.WriteString(fmt.Sprintf("| %s |\n", rule))
		}
		buf.WriteString("\n")
	} else {
		buf.WriteString("未发现有效规则。\n\n")
	}

	// 测试建议
	buf.WriteString("## 测试建议\n\n")
	buf.WriteString("### 下一步操作\n\n")
	buf.WriteString("1. **修复错误规则**: 针对编译错误的规则，检查并修复正则表达式或 SampleCode\n")
	buf.WriteString("2. **补充 SampleCode**: 为缺少 SampleCode 的规则添加合适的测试样本\n")
	buf.WriteString("3. **优化规则**: 分析有效规则的性能和准确性，进行必要的优化\n")
	buf.WriteString("4. **定期测试**: 建立规则测试的定期执行机制，确保规则的持续有效性\n\n")

	// 报告尾部
	buf.WriteString("## 报告信息\n\n")
	buf.WriteString(fmt.Sprintf("- **报告生成时间**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	buf.WriteString("- **报告版本**: 1.0.0\n")
	buf.WriteString("- **生成工具**: PrivacyCheckGo\n")

	return buf.String()
}

// getHostname 获取主机名
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
