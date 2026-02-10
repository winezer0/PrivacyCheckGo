package baserule

import (
	"fmt"
	"regexp"
	"time"

	"github.com/dlclark/regexp2"
)

// RegexEngine 定义支持的正则引擎类型
type RegexEngine string

const (
	RegexEngineGo   RegexEngine = "go"   // Go 标准库 RE2 引擎
	RegexEngineJava RegexEngine = "java" // Java 风格正则（使用 regexp2）
)

// RegexMatcher 定义通用正则匹配接口
type RegexMatcher interface {
	MatchString(s string) (bool, error)
	FindStringMatch(s string) (MatchResult, error)
	FindAllString(s string, n int) []string
}

// MatchResult 定义匹配结果接口
type MatchResult interface {
	String() string
	Groups() []Group
	FindNextMatch() (MatchResult, error)
}

// Group 定义分组接口
type Group interface {
	String() string
	Captures() []Capture
}

// Capture 定义捕获接口
type Capture interface {
	String() string
}

// GoRegexMatcher 实现 Go 标准库正则的匹配器
type GoRegexMatcher struct {
	regex *regexp.Regexp
}

// NewGoRegexMatcher 创建 Go 正则匹配器
func NewGoRegexMatcher(pattern string) (*GoRegexMatcher, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &GoRegexMatcher{regex: regex}, nil
}

// MatchString 匹配字符串
func (m *GoRegexMatcher) MatchString(s string) (bool, error) {
	return m.regex.MatchString(s), nil
}

// FindStringMatch 查找第一个匹配
func (m *GoRegexMatcher) FindStringMatch(s string) (MatchResult, error) {
	index := m.regex.FindStringIndex(s)
	if index == nil {
		return nil, nil
	}
	return &GoMatchResult{
		s:     s,
		start: index[0],
		end:   index[1],
		regex: m.regex,
	}, nil
}

// FindAllString 查找所有匹配
func (m *GoRegexMatcher) FindAllString(s string, n int) []string {
	return m.regex.FindAllString(s, n)
}

// GoMatchResult 实现 Go 标准库的匹配结果
type GoMatchResult struct {
	s     string
	start int
	end   int
	regex *regexp.Regexp
}

// String 返回匹配的字符串
func (r *GoMatchResult) String() string {
	if r.start < 0 || r.end > len(r.s) {
		return ""
	}
	return r.s[r.start:r.end]
}

// Groups 返回分组信息
func (r *GoMatchResult) Groups() []Group {
	return []Group{&GoGroup{value: r.String()}}
}

// FindNextMatch 查找下一个匹配
func (r *GoMatchResult) FindNextMatch() (MatchResult, error) {
	if r.end >= len(r.s) {
		return nil, nil
	}
	index := r.regex.FindStringIndex(r.s[r.end:])
	if index == nil {
		return nil, nil
	}
	return &GoMatchResult{
		s:     r.s,
		start: r.end + index[0],
		end:   r.end + index[1],
		regex: r.regex,
	}, nil
}

// GoGroup 实现 Go 标准库的分组
type GoGroup struct {
	value string
}

// String 返回分组值
func (g *GoGroup) String() string {
	return g.value
}

// Captures 返回捕获信息
func (g *GoGroup) Captures() []Capture {
	return []Capture{&GoCapture{value: g.value}}
}

// GoCapture 实现 Go 标准库的捕获
type GoCapture struct {
	value string
}

// String 返回捕获值
func (c *GoCapture) String() string {
	return c.value
}

// JavaRegexMatcher 实现 Java 风格正则的匹配器
type JavaRegexMatcher struct {
	regex *regexp2.Regexp
}

// NewJavaRegexMatcher 创建 Java 正则匹配器
func NewJavaRegexMatcher(pattern string) (*JavaRegexMatcher, error) {
	regex, err := regexp2.Compile(pattern, regexp2.None)
	if err != nil {
		return nil, err
	}
	// 设置超时，防止灾难性回溯
	regex.MatchTimeout = 5 * time.Second
	return &JavaRegexMatcher{regex: regex}, nil
}

// MatchString 匹配字符串
func (m *JavaRegexMatcher) MatchString(s string) (bool, error) {
	return m.regex.MatchString(s)
}

// FindStringMatch 查找第一个匹配
func (m *JavaRegexMatcher) FindStringMatch(s string) (MatchResult, error) {
	match, err := m.regex.FindStringMatch(s)
	if err != nil {
		return nil, err
	}
	if match == nil {
		return nil, nil
	}
	return &JavaMatchResult{match: match, regex: m.regex}, nil
}

// FindAllString 查找所有匹配
func (m *JavaRegexMatcher) FindAllString(s string, n int) []string {
	var matches []string
	match, err := m.regex.FindStringMatch(s)
	if err != nil || match == nil {
		return matches
	}

	count := 0
	for match != nil && (n <= 0 || count < n) {
		matches = append(matches, match.String())
		match, _ = m.regex.FindNextMatch(match)
		count++
	}
	return matches
}

// JavaMatchResult 实现 regexp2 的匹配结果
type JavaMatchResult struct {
	match *regexp2.Match
	regex *regexp2.Regexp
}

// String 返回匹配的字符串
func (r *JavaMatchResult) String() string {
	if r.match == nil {
		return ""
	}
	return r.match.String()
}

// Groups 返回分组信息
func (r *JavaMatchResult) Groups() []Group {
	if r.match == nil {
		return []Group{}
	}
	var groups []Group
	for i := 0; i < r.match.GroupCount(); i++ {
		group := r.match.GroupByNumber(i)
		if group != nil {
			groups = append(groups, &JavaGroup{group: group})
		}
	}
	return groups
}

// FindNextMatch 查找下一个匹配
func (r *JavaMatchResult) FindNextMatch() (MatchResult, error) {
	if r.match == nil {
		return nil, nil
	}
	match, err := r.regex.FindNextMatch(r.match)
	if err != nil || match == nil {
		return nil, err
	}
	return &JavaMatchResult{match: match, regex: r.regex}, nil
}

// JavaGroup 实现 regexp2 的分组
type JavaGroup struct {
	group *regexp2.Group
}

// String 返回分组值
func (g *JavaGroup) String() string {
	if g.group == nil {
		return ""
	}
	return g.group.String()
}

// Captures 返回捕获信息
func (g *JavaGroup) Captures() []Capture {
	if g.group == nil {
		return []Capture{}
	}
	var captures []Capture
	for _, capture := range g.group.Captures {
		captures = append(captures, &JavaCapture{capture: capture})
	}
	return captures
}

// JavaCapture 实现 regexp2 的捕获
type JavaCapture struct {
	capture regexp2.Capture
}

// String 返回捕获值
func (c *JavaCapture) String() string {
	return c.capture.String()
}

// NewRegexMatcher 根据引擎类型创建正则匹配器
func NewRegexMatcher(pattern string, engine RegexEngine) (RegexMatcher, error) {
	switch engine {
	case RegexEngineGo:
		return NewGoRegexMatcher(pattern)
	case RegexEngineJava:
		return NewJavaRegexMatcher(pattern)
	default:
		return nil, fmt.Errorf("unsupported regex engine: %s", engine)
	}
}

// TryCompileWithFallback 尝试使用 Go 引擎编译，失败则使用 Java 引擎
func TryCompileWithFallback(pattern string) (RegexMatcher, error) {
	// 首先尝试使用 Go 标准库引擎
	goMatcher, err := NewGoRegexMatcher(pattern)
	if err == nil {
		return goMatcher, nil
	}

	// Go 引擎编译失败，尝试使用 Java 引擎（regexp2）
	javaMatcher, err := NewJavaRegexMatcher(pattern)
	if err == nil {
		return javaMatcher, nil
	}

	// 两种引擎都失败
	return nil, fmt.Errorf("both engines failed to compile regex: %v", err)
}
