package scanner

import (
	"testing"

	"privacycheck/internal/baserule"
)

// TestRuleEngineWithGoEngine 测试使用 Go 引擎的规则
func TestRuleEngineWithGoEngine(t *testing.T) {
	// 创建测试规则
	rules := baserule.RuleMap{
		"test": {
			{
				Name:    "Email",
				FRegex:  "\\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b",
				Engine:  "go",
				Loaded:  true,
				Sensitive: true,
			},
		},
	}

	// 创建规则引擎
	engine, err := NewRuleEngine(rules)
	if err != nil {
		t.Fatalf("NewRuleEngine failed: %v", err)
	}

	// 测试匹配
	content := "Contact us at test@example.com for more information"
	results := engine.ApplyRules(content, "test.txt", 0, 1)

	if len(results) == 0 {
		t.Errorf("Expected at least one match, but got none")
	}

	// 验证匹配结果
	found := false
	for _, result := range results {
		if result.Match == "test@example.com" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find 'test@example.com', but didn't")
	}
}

// TestRuleEngineWithJavaEngine 测试使用 Java 引擎的规则
func TestRuleEngineWithJavaEngine(t *testing.T) {
	// 创建测试规则
	rules := baserule.RuleMap{
		"test": {
			{
				Name:    "Email",
				FRegex:  "\\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b",
				Engine:  "java",
				Loaded:  true,
				Sensitive: true,
			},
		},
	}

	// 创建规则引擎
	engine, err := NewRuleEngine(rules)
	if err != nil {
		t.Fatalf("NewRuleEngine failed: %v", err)
	}

	// 测试匹配
	content := "Contact us at test@example.com for more information"
	results := engine.ApplyRules(content, "test.txt", 0, 1)

	if len(results) == 0 {
		t.Errorf("Expected at least one match, but got none")
	}

	// 验证匹配结果
	found := false
	for _, result := range results {
		if result.Match == "test@example.com" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find 'test@example.com', but didn't")
	}
}

// TestRuleEngineWithEmptyEngine 测试引擎为空时的回退机制
func TestRuleEngineWithEmptyEngine(t *testing.T) {
	// 创建测试规则（引擎为空）
	rules := baserule.RuleMap{
		"test": {
			{
				Name:    "Email",
				FRegex:  "\\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b",
				Engine:  "", // 引擎为空
				Loaded:  true,
				Sensitive: true,
			},
		},
	}

	// 创建规则引擎
	engine, err := NewRuleEngine(rules)
	if err != nil {
		t.Fatalf("NewRuleEngine failed: %v", err)
	}

	// 测试匹配
	content := "Contact us at test@example.com for more information"
	results := engine.ApplyRules(content, "test.txt", 0, 1)

	if len(results) == 0 {
		t.Errorf("Expected at least one match, but got none")
	}

	// 验证匹配结果
	found := false
	for _, result := range results {
		if result.Match == "test@example.com" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find 'test@example.com', but didn't")
	}
}

// TestRuleEngineWithMultipleRules 测试多个规则
func TestRuleEngineWithMultipleRules(t *testing.T) {
	// 创建测试规则
	rules := baserule.RuleMap{
		"test": {
			{
				Name:    "Email",
				FRegex:  "\\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,}\\b",
				Engine:  "", // 引擎为空，使用回退机制
				Loaded:  true,
				Sensitive: true,
			},
			{
				Name:    "Phone",
				FRegex:  "\\b\\d{3}-\\d{3}-\\d{4}\\b",
				Engine:  "go", // 使用 Go 引擎
				Loaded:  true,
				Sensitive: true,
			},
		},
	}

	// 创建规则引擎
	engine, err := NewRuleEngine(rules)
	if err != nil {
		t.Fatalf("NewRuleEngine failed: %v", err)
	}

	// 测试匹配
	content := "Contact us at test@example.com or call 123-456-7890"
	results := engine.ApplyRules(content, "test.txt", 0, 1)

	if len(results) < 2 {
		t.Errorf("Expected at least two matches, but got %d", len(results))
	}

	// 验证匹配结果
	foundEmail := false
	foundPhone := false
	for _, result := range results {
		if result.Match == "test@example.com" {
			foundEmail = true
		} else if result.Match == "123-456-7890" {
			foundPhone = true
		}
	}

	if !foundEmail {
		t.Errorf("Expected to find 'test@example.com', but didn't")
	}

	if !foundPhone {
		t.Errorf("Expected to find '123-456-7890', but didn't")
	}
}
