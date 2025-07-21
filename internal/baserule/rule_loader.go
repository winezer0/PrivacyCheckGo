package baserule

import (
	"fmt"
	"privacycheck/internal/embeds"
	"privacycheck/pkg/fileutils"

	"gopkg.in/yaml.v3"
)

// LoadRulesYaml 加载规则配置文件
func LoadRulesYaml(configPath string) (*RuleConfig, error) {
	// 读取配置文件
	data, err := fileutils.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the rule file:%s error: %w", configPath, err)
	}

	var ruleConfig RuleConfig

	if err := yaml.Unmarshal(data, &ruleConfig); err != nil {
		return nil, fmt.Errorf("failed to parse the rule file:%s error: %w", configPath, err)
	}
	return &ruleConfig, nil
}

// CreateDefaultConfig 创建默认配置文件
func CreateDefaultConfig(configPath string) error {
	// 写入默认配置（fileutils.WriteFile会自动创建目录）
	if err := fileutils.WriteFile(configPath, embeds.DefaultConfig); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	return nil
}
