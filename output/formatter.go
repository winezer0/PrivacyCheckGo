package output

import (
	"privacycheck/scanner"
	"strings"
)

// formatResults 格式化结果
func (p *Output) formatResults(results []scanner.ScanResult) []scanner.ScanResult {
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
