package baserule

// Rule 表示单个检测规则
type Rule struct {
	Name         string `yaml:"name" json:"name"`                   // 规则名称
	FRegex       string `yaml:"f_regex" json:"f_regex"`             // 正则表达式
	SRegex       string `yaml:"s_regex" json:"s_regex"`             // 二次匹配规则(未实现)
	Format       string `yaml:"format" json:"format"`               // 结果提取格式(未实现)
	Color        string `yaml:"color" json:"color"`                 // 结果颜色显示(未实现)
	Scope        string `yaml:"scope" json:"scope"`                 // 规则匹配范围(未实现)
	Engine       string `yaml:"engine" json:"engine"`               // 规则匹配引擎(未实现)
	Sensitive    bool   `yaml:"sensitive" json:"sensitive"`         // 是否敏感信息
	Loaded       bool   `yaml:"loaded" json:"loaded"`               // 是否启用规则
	ContextLeft  int    `yaml:"context_left" json:"context_left"`   // 匹配结果向左扩充字符数
	ContextRight int    `yaml:"context_right" json:"context_right"` // 匹配结果向右扩充字符数
}

// Rules 表示规则组
type Rules struct {
	Group string `yaml:"group" json:"group"` // 规则组名称
	Rule  []Rule `yaml:"rule" json:"rule"`   // 规则列表
}

// RuleConfig 表示完整的规则配置
type RuleConfig struct {
	Rules []Rules `yaml:"rules" json:"rules"`
}
