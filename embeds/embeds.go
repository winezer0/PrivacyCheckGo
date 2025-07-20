package embeds

import (
	_ "embed"
)

// DefaultConfig 默认配置文件内容
//go:embed config.yaml
var DefaultConfig []byte
