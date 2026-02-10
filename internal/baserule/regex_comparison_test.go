package baserule

import (
	"testing"
)

// TestRegexEnginesComparison 测试不同引擎的正则表达式匹配
func TestRegexEnginesComparison(t *testing.T) {
	testCases := []struct {
		name        string
		pattern     string
		engine      RegexEngine
		testString  string
		expected    bool
		description string
	}{
		// 基础正则表达式 - 所有引擎都应该支持
		{
			name:        "Basic Pattern - Go Engine",
			pattern:     "\\d{3}-\\d{3}-\\d{4}",
			engine:      RegexEngineGo,
			testString:  "My phone number is 123-456-7890",
			expected:    true,
			description: "Basic digit pattern with Go engine",
		},
		{
			name:        "Basic Pattern - Java Engine",
			pattern:     "\\d{3}-\\d{3}-\\d{4}",
			engine:      RegexEngineJava,
			testString:  "My phone number is 123-456-7890",
			expected:    true,
			description: "Basic digit pattern with Java engine",
		},

		// 分组命名 - Java和Go的差异
		{
			name:        "Named Group - Java Style",
			pattern:     "(?<year>\\d{4})-(?<month>\\d{2})-(?<day>\\d{2})",
			engine:      RegexEngineJava,
			testString:  "Today is 2023-12-25",
			expected:    true,
			description: "Java style named groups",
		},
		{
			name:        "Named Group - Go Style",
			pattern:     "(?P<year>\\d{4})-(?P<month>\\d{2})-(?P<day>\\d{2})",
			engine:      RegexEngineGo,
			testString:  "Today is 2023-12-25",
			expected:    true,
			description: "Go style named groups",
		},

		// 前瞻断言 - Go标准库不支持，但regexp2支持
		{
			name:        "Positive Lookahead - Java Engine",
			pattern:     "\\d+(?= dollars)",
			engine:      RegexEngineJava,
			testString:  "The price is 100 dollars",
			expected:    true,
			description: "Positive lookahead with Java engine",
		},

		// 后顾断言 - Go标准库不支持，但regexp2支持
		{
			name:        "Positive Lookbehind - Java Engine",
			pattern:     "(?<=\\$)\\d+",
			engine:      RegexEngineJava,
			testString:  "The price is $100",
			expected:    true,
			description: "Positive lookbehind with Java engine",
		},

		// 反向引用
		{
			name:        "Back Reference - Java Engine",
			pattern:     "(\\w+) \\1",
			engine:      RegexEngineJava,
			testString:  "hello hello",
			expected:    true,
			description: "Back reference with Java engine",
		},

		// 命名反向引用
		{
			name:        "Named Back Reference - Java Engine",
			pattern:     "(?<word>\\w+) \\k'word'",
			engine:      RegexEngineJava,
			testString:  "hello hello",
			expected:    true,
			description: "Named back reference with Java engine",
		},

		// 条件表达式
		{
			name:        "Conditional - Java Engine",
			pattern:     "(\\d+)(?(1) dollars| euro)",
			engine:      RegexEngineJava,
			testString:  "100 dollars",
			expected:    true,
			description: "Conditional regex with Java engine",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 编译正则表达式
			matcher, err := NewRegexMatcher(tc.pattern, tc.engine)
			if err != nil {
				t.Fatalf("Failed to create matcher for %s: %v", tc.description, err)
			}

			// 测试匹配
			actual, err := matcher.MatchString(tc.testString)
			if err != nil {
				t.Fatalf("Error matching for %s: %v", tc.description, err)
			}

			if actual != tc.expected {
				t.Errorf("%s: expected %v, got %v", tc.description, tc.expected, actual)
			}
		})
	}
}

// TestTryCompileWithFallback 测试空引擎名称的回退机制
func TestTryCompileWithFallback_Advanced(t *testing.T) {
	testCases := []struct {
		name        string
		pattern     string
		testString  string
		expected    bool
		description string
	}{
		// Go引擎可以处理的正则
		{
			name:        "Go Compatible Pattern",
			pattern:     "(?i)\\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b",
			testString:  "Contact at test@example.com",
			expected:    true,
			description: "Email pattern compatible with Go engine",
		},

		// 包含Java特有语法的正则（应该回退到Java引擎）
		{
			name:        "Java Specific Pattern",
			pattern:     "\\d+(?= dollars)",
			testString:  "Price is 50 dollars",
			expected:    true,
			description: "Pattern with lookahead, should use Java engine",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 使用回退机制编译
			matcher, err := TryCompileWithFallback(tc.pattern)
			if err != nil {
				t.Fatalf("Failed to compile with fallback for %s: %v", tc.description, err)
			}

			// 测试匹配
			actual, err := matcher.MatchString(tc.testString)
			if err != nil {
				t.Fatalf("Error matching for %s: %v", tc.description, err)
			}

			if actual != tc.expected {
				t.Errorf("%s: expected %v, got %v", tc.description, tc.expected, actual)
			}
		})
	}
}

// TestEngineSpecificFeatures 测试引擎特定的功能
func TestEngineSpecificFeatures(t *testing.T) {
	// 测试Java引擎特有的功能
	javaOnlyPatterns := []struct {
		pattern     string
		testString  string
		expected    bool
		description string
	}{
		// 前瞻断言
		{"\\d+(?= euros)", "Price is 100 euros", true, "Positive lookahead"},
		// 后顾断言
		{"(?<=\\$)\\d+", "$50", true, "Positive lookbehind"},
		// 负前瞻
		{"\\d+(?! dollars)", "Price is 100 euros", true, "Negative lookahead"},
		// 负后顾
		{"(?<!\\$)\\d+", "Price is 100", true, "Negative lookbehind"},
		// 条件表达式
		{"(\\d+)(?(1) euros|)", "100 euros", true, "Conditional expression"},
	}

	// 测试Java引擎特有功能
	for _, tc := range javaOnlyPatterns {
		t.Run("Java Engine - "+tc.description, func(t *testing.T) {
			matcher, err := NewJavaRegexMatcher(tc.pattern)
			if err != nil {
				t.Fatalf("Failed to create Java matcher for %s: %v", tc.description, err)
			}

			actual, err := matcher.MatchString(tc.testString)
			if err != nil {
				t.Fatalf("Error matching for %s: %v", tc.description, err)
			}

			if actual != tc.expected {
				t.Errorf("%s: expected %v, got %v", tc.description, tc.expected, actual)
			}
		})
	}

	// 测试Go引擎的限制
	goLimitedPatterns := []struct {
		pattern     string
		description string
	}{
		{"\\d+(?= euros)", "Positive lookahead (Go may not support)"},
		{"(?<=\\$)\\d+", "Positive lookbehind (Go may not support)"},
		{"\\d+(?! dollars)", "Negative lookahead (Go may not support)"},
		{"(?<!\\$)\\d+", "Negative lookbehind (Go may not support)"},
	}

	// 测试Go引擎对高级特性的处理
	for _, tc := range goLimitedPatterns {
		t.Run("Go Engine - "+tc.description, func(t *testing.T) {
			_, err := NewGoRegexMatcher(tc.pattern)
			// 这里我们不期望失败，因为我们使用的是基本语法
			// 但如果使用了Go不支持的高级特性，可能会失败
			if err != nil {
				t.Logf("Go engine failed to compile %s: %v", tc.description, err)
				// 注意：这不是错误，只是记录Go引擎的限制
			}
		})
	}
}

// TestRegexEnginesPerformance 简单的性能比较测试
func TestRegexEnginesPerformance(t *testing.T) {
	// 这个测试主要是为了确保引擎不会在简单模式下崩溃
	// 实际性能测试应该使用更复杂的基准测试

	simplePattern := "\\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b"
	testString := "Contact us at test@example.com for more information"

	// 测试Go引擎
	goMatcher, err := NewGoRegexMatcher(simplePattern)
	if err != nil {
		t.Fatalf("Failed to create Go matcher: %v", err)
	}

	for i := 0; i < 100; i++ {
		_, err := goMatcher.MatchString(testString)
		if err != nil {
			t.Fatalf("Go engine failed during performance test: %v", err)
		}
	}

	// 测试Java引擎
	javaMatcher, err := NewJavaRegexMatcher(simplePattern)
	if err != nil {
		t.Fatalf("Failed to create Java matcher: %v", err)
	}

	for i := 0; i < 100; i++ {
		_, err := javaMatcher.MatchString(testString)
		if err != nil {
			t.Fatalf("Java engine failed during performance test: %v", err)
		}
	}
}
