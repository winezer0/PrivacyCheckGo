package baserule

import (
	"fmt"
	"github.com/winezer0/xutils/utils"
	"privacycheck/internal/embeds"
)

// LoadRulesYaml 加载规则配置文件
func LoadRulesYaml(configPath string) (*RuleConfig, error) {
	// 读取配置文件
	var ruleConfig RuleConfig
	if err := utils.LoadYAML(configPath, &ruleConfig); err != nil {
		return nil, fmt.Errorf("failed to parse the rule file:%s error: %w", configPath, err)
	}

	return &ruleConfig, nil
}

// CreateDefaultConfig 创建默认配置文件
func CreateDefaultConfig(configPath string) error {
	// 写入默认配置（fileutils.WriteFile会自动创建目录）
	if err := utils.SaveToFile(configPath, embeds.DefaultConfig); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}
	return nil
}
