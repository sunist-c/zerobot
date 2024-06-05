package manager

type PluginMetadata struct {
	Name         string            `json:"name" yaml:"name"`                             // 插件的名称
	Group        string            `json:"group,omitempty" yaml:"group"`                 // 插件的组名
	Description  string            `json:"description,omitempty" yaml:"description"`     // 插件的描述信息
	Help         string            `json:"help,omitempty" yaml:"help"`                   // 插件的帮助信息
	Banner       string            `json:"banner,omitempty" yaml:"banner"`               // 插件的横幅图片
	DataFolder   string            `json:"data_folder,omitempty" yaml:"data_folder"`     // 插件的数据文件夹
	PublicFolder string            `json:"public_folder,omitempty" yaml:"public_folder"` // 插件的公共数据文件夹
	Enable       bool              `json:"enable,omitempty" yaml:"enable"`               // 插件是否启用
	Handlers     []HandlerMetadata `json:"handlers,omitempty" yaml:"handlers"`           // 插件的处理器
}

type HandlerMetadata struct {
	HandlerName string                    `json:"name" yaml:"name"`                         // 插件的处理器名称
	Group       string                    `json:"group,omitempty" yaml:"group"`             // 插件的组名
	Blocked     bool                      `json:"blocked,omitempty" yaml:"blocked"`         // 插件是否会阻塞
	Limiter     string                    `json:"limiter,omitempty" yaml:"limiter"`         // 插件的限流器
	Middlewares HandlerMiddlewareMetadata `json:"middlewares,omitempty" yaml:"middlewares"` // 插件的中间件
	Triggers    HandlerTriggerMetadata    `json:"triggers,omitempty" yaml:"triggers"`       // 插件的触发器
}

type HandlerMiddlewareMetadata struct {
	PreHandlers []string `json:"pre_handlers,omitempty" yaml:"pre_handlers"` // 插件的前置处理器，zero.UsePreHandler
	MidHandlers []string `json:"mid_handlers,omitempty" yaml:"mid_handlers"` // 插件的中间处理器，zero.UseMidHandler
}

type HandlerTriggerMetadata struct {
	FullMatches []string `json:"full_matches,omitempty" yaml:"full_matches"` // 插件的完全匹配，zero.OnFullMatchGroup
	Keywords    []string `json:"keywords,omitempty" yaml:"keywords"`         // 插件的关键词匹配，zero.OnKeywordGroup
	Commands    []string `json:"commands,omitempty" yaml:"commands"`         // 插件的触发命令，zero.OnCommandGroup
	Prefixes    []string `json:"prefixes,omitempty" yaml:"prefixes"`         // 插件的触发前缀，zero.OnPrefixGroup
	Suffixes    []string `json:"suffixes,omitempty" yaml:"suffixes"`         // 插件的触发后缀，zero.OnSuffixGroup
	Regexes     []string `json:"regexes,omitempty" yaml:"regexes"`           // 插件的正则匹配，zero.OnRegex
	Notice      bool     `json:"notice,omitempty" yaml:"notice"`             // 插件的通知匹配，zero.OnNotice
}

type PluginGroup struct {
	Name        string        `json:"name" yaml:"name"`                           // 插件组的名称
	PreHandlers []string      `json:"pre_handlers,omitempty" yaml:"pre_handlers"` // 插件组的前置处理器，zero.UsePreHandler
	MidHandlers []string      `json:"mid_handlers,omitempty" yaml:"mid_handlers"` // 插件组的中间处理器，zero.UseMidHandler
	SubGroups   []PluginGroup `json:"sub_groups,omitempty" yaml:"sub_groups"`     // 插件组的子插件组
}

type ManagedConfig struct {
	Groups  []PluginGroup    `json:"groups" yaml:"groups"`   // 插件组
	Plugins []PluginMetadata `json:"plugins" yaml:"plugins"` // 插件
}
