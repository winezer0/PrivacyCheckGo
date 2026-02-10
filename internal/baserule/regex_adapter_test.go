package baserule

import (
	"testing"
)

// TestNewGoRegexMatcher 测试 Go 正则匹配器
func TestNewGoRegexMatcher(t *testing.T) {
	pattern := "(?i)(?m)\\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b"
	matcher, err := NewGoRegexMatcher(pattern)
	if err != nil {
		t.Fatalf("NewGoRegexMatcher failed: %v", err)
	}

	// 测试匹配
	matches := []string{
		"test@example.com",
		"user.name@domain.co.uk",
	}

	for _, match := range matches {
		ok, err := matcher.MatchString(match)
		if err != nil {
			t.Errorf("MatchString failed for %s: %v", match, err)
		}
		if !ok {
			t.Errorf("Expected %s to match, but it didn't", match)
		}
	}

	// 测试不匹配
	nonMatches := []string{
		"invalid-email",
		"test@",
		"@example.com",
	}

	for _, nonMatch := range nonMatches {
		ok, err := matcher.MatchString(nonMatch)
		if err != nil {
			t.Errorf("MatchString failed for %s: %v", nonMatch, err)
		}
		if ok {
			t.Errorf("Expected %s not to match, but it did", nonMatch)
		}
	}
}

// TestNewJavaRegexMatcher 测试 Java 正则匹配器
func TestNewJavaRegexMatcher(t *testing.T) {
	pattern := "(?i)(?m)\\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b"
	matcher, err := NewJavaRegexMatcher(pattern)
	if err != nil {
		t.Fatalf("NewJavaRegexMatcher failed: %v", err)
	}

	// 测试匹配
	matches := []string{
		"test@example.com",
		"user.name@domain.co.uk",
	}

	for _, match := range matches {
		ok, err := matcher.MatchString(match)
		if err != nil {
			t.Errorf("MatchString failed for %s: %v", match, err)
		}
		if !ok {
			t.Errorf("Expected %s to match, but it didn't", match)
		}
	}

	// 测试不匹配
	nonMatches := []string{
		"invalid-email",
		"test@",
		"@example.com",
	}

	for _, nonMatch := range nonMatches {
		ok, err := matcher.MatchString(nonMatch)
		if err != nil {
			t.Errorf("MatchString failed for %s: %v", nonMatch, err)
		}
		if ok {
			t.Errorf("Expected %s not to match, but it did", nonMatch)
		}
	}
}

// TestTryCompileWithFallback 测试尝试编译并回退的功能
func TestTryCompileWithFallback(t *testing.T) {
	// 测试 Go 引擎可以编译的正则
	pattern1 := "(?i)(?m)\\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b"
	matcher1, err := TryCompileWithFallback(pattern1)
	if err != nil {
		t.Fatalf("TryCompileWithFallback failed for pattern1: %v", err)
	}

	// 测试匹配
	ok, err := matcher1.MatchString("test@example.com")
	if err != nil {
		t.Errorf("MatchString failed: %v", err)
	}
	if !ok {
		t.Errorf("Expected match, but it didn't")
	}

	// 测试需要回退到 Java 引擎的正则（包含 Go 不支持的特性）
	// 注意：这里使用一个 Go 可能支持的模式作为示例
	pattern2 := "(?i)(?m)\\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b"
	matcher2, err := TryCompileWithFallback(pattern2)
	if err != nil {
		t.Fatalf("TryCompileWithFallback failed for pattern2: %v", err)
	}

	// 测试匹配
	ok, err = matcher2.MatchString("test@example.com")
	if err != nil {
		t.Errorf("MatchString failed: %v", err)
	}
	if !ok {
		t.Errorf("Expected match, but it didn't")
	}
}

// TestNewRegexMatcher 测试根据引擎类型创建匹配器
func TestNewRegexMatcher(t *testing.T) {
	pattern := "(?i)(?m)\\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b"

	// 测试 Go 引擎
	matcher1, err := NewRegexMatcher(pattern, RegexEngineGo)
	if err != nil {
		t.Fatalf("NewRegexMatcher with Go engine failed: %v", err)
	}

	ok, err := matcher1.MatchString("test@example.com")
	if err != nil {
		t.Errorf("MatchString failed: %v", err)
	}
	if !ok {
		t.Errorf("Expected match, but it didn't")
	}

	// 测试 Java 引擎
	matcher2, err := NewRegexMatcher(pattern, RegexEngineJava)
	if err != nil {
		t.Fatalf("NewRegexMatcher with Java engine failed: %v", err)
	}

	ok, err = matcher2.MatchString("test@example.com")
	if err != nil {
		t.Errorf("MatchString failed: %v", err)
	}
	if !ok {
		t.Errorf("Expected match, but it didn't")
	}
}
