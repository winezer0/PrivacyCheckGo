package config

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"privacycheck/baserule"
	"privacycheck/fileutils"
	"privacycheck/internal/embeds"
	"privacycheck/pkg/logging"
	"strings"
)

// LoadRulesConfig 加载规则配置文件
func LoadRulesConfig(configPath string) (*baserule.RulesConfig, error) {
	// 如果配置文件不存在，创建默认配置文件
	if !fileutils.FileExists(configPath) {
		logging.Infof("配置文件 %s 不存在，正在创建默认配置文件...", configPath)
		if err := createDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %w", err)
		}
		logging.Infof("默认配置文件已创建: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config baserule.RulesConfig

	// 根据文件扩展名决定解析方式
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("解析JSON配置文件失败: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("解析YAML配置文件失败: %w", err)
		}
	default:
		// 默认尝试YAML解析
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("解析配置文件失败（默认YAML格式）: %w", err)
		}
	}

	// 兼容HAE格式：如果配置文件包含rules节点，提取rules内容
	if len(config.Rules) == 0 {
		// 尝试解析为包含rules节点的格式
		var haeConfig struct {
			Rules []baserule.RuleGroup `yaml:"rules" json:"rules"`
		}

		switch ext {
		case ".json":
			if err := json.Unmarshal(data, &haeConfig); err == nil && len(haeConfig.Rules) > 0 {
				config.Rules = haeConfig.Rules
			}
		default:
			if err := yaml.Unmarshal(data, &haeConfig); err == nil && len(haeConfig.Rules) > 0 {
				config.Rules = haeConfig.Rules
			}
		}
	}

	return &config, nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 写入默认配置
	if err := os.WriteFile(configPath, embeds.DefaultConfig, 0644); err != nil {
		return fmt.Errorf("写入默认配置文件失败: %w", err)
	}

	return nil
}
