package baserule

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"privacycheck/internal/embeds"
)

// LoadRulesYaml 加载规则配置文件
func LoadRulesYaml(configPath string) (*RuleConfig, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the rule file:%s error: %w", configPath, err)
	}

	var config RuleConfig

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse the rule file:%s error: %w", configPath, err)
	}
	return &config, nil
}

// CreateDefaultConfig 创建默认配置文件
func CreateDefaultConfig(configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	// 写入默认配置
	if err := os.WriteFile(configPath, embeds.DefaultConfig, 0644); err != nil {
		return fmt.Errorf("failed to writee  dfault config file: %w", err)
	}

	return nil
}
