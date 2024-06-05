package manager

import (
	"embed"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/alioth-center/infrastructure/config"
	"github.com/alioth-center/infrastructure/logger"
	"github.com/alioth-center/infrastructure/utils/concurrency"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
)

var (
	logging logger.Logger

	//go:embed embedding/*
	embeddingConfig embed.FS

	manager = &pluginManager{
		handlers:    concurrency.NewHashMap[string, zero.Handler](concurrency.HashMapNodeOptionSmallSize),
		middlewares: concurrency.NewHashMap[string, zero.Rule](concurrency.HashMapNodeOptionSmallSize),
		groups:      concurrency.NewHashMap[string, *HandlerMiddlewareMetadata](concurrency.HashMapNodeOptionSmallSize),
	}
)

func SetLogger(log logger.Logger) {
	if log != nil && logging == nil {
		logging = log
	}
}

func Default() Manager {
	return manager
}

func Initialize(externalConfig string) {
	if logging == nil {
		SetLogger(logger.Default())
	}

	// read embedding config
	var yamlConfig, jsonConfig ManagedConfig
	yamlReadErr, jsonReadErr :=
		config.LoadEmbedConfig(&yamlConfig, embeddingConfig, "embeddings/config.yaml"),
		config.LoadEmbedConfig(&jsonConfig, embeddingConfig, "embeddings/config.json")
	if yamlReadErr != nil && jsonReadErr != nil {
		logging.Info(logger.NewFields().WithMessage("failed to read embedding config").WithData(map[string]string{"yaml": yamlReadErr.Error(), "json": jsonReadErr.Error()}))
	}

	// read external config
	var external ManagedConfig
	externalReadErr := config.LoadConfig(&external, externalConfig)
	if externalReadErr != nil {
		logging.Error(logger.NewFields().WithMessage("failed to read external config").WithData(map[string]string{"path": externalConfig, "error": externalReadErr.Error()}))
	}

	// merge configs
	manager.InitializeManagedPlugins(&yamlConfig)
	manager.InitializeManagedPlugins(&jsonConfig)
	manager.InitializeManagedPlugins(&external)
}

type Manager interface {
	RegisterHandler(name string, handler zero.Handler)
	RegisterMiddleware(name string, middleware zero.Rule)
}

type pluginManager struct {
	handlers    concurrency.Map[string, zero.Handler]
	middlewares concurrency.Map[string, zero.Rule]
	groups      concurrency.Map[string, *HandlerMiddlewareMetadata]
}

func (p *pluginManager) RegisterHandler(name string, handler zero.Handler) {
	p.handlers.Set(name, handler)
}

func (p *pluginManager) RegisterMiddleware(name string, middleware zero.Rule) {
	p.middlewares.Set(name, middleware)
}

func (p *pluginManager) InitializeManagedPlugins(conf *ManagedConfig) {
	p.parseGroups(conf.Groups)

	for _, plugin := range conf.Plugins {
		if !plugin.Enable || len(plugin.Handlers) == 0 {
			continue
		}

		endpoints := map[string]zero.Handler{}
		for _, handler := range plugin.Handlers {
			impl, got := p.handlers.Get(handler.HandlerName)
			if got && impl != nil {
				endpoints[handler.HandlerName] = impl
			}
		}
		if len(endpoints) == 0 {
			continue
		}

		engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
			DisableOnDefault:  false,
			Brief:             plugin.Description,
			Help:              plugin.Help,
			Banner:            plugin.Banner,
			PrivateDataFolder: plugin.DataFolder,
			PublicDataFolder:  plugin.PublicFolder,
			OnEnable:          pluginEnableLoggingCallback(plugin),
			OnDisable:         pluginDisableLoggingCallback(plugin),
		})

		pluginPreHandlers, pluginMidHandlers := p.getGroupPreHandlers(plugin.Group), p.getGroupMidHandlers(plugin.Group)
		for _, handler := range plugin.Handlers {
			if endpoints[handler.HandlerName] == nil {
				continue
			}

			// middlewares priority:
			// 1. handler.Middlewares[must]
			// 2. handler.Group.Middlewares[if not empty]
			// 3. plugin.Group.Middlewares[if not empty && handler.Group is empty]
			handlerPreHandlers, handlerMidHandlers := p.getGroupPreHandlers(handler.Group), p.getGroupMidHandlers(handler.Group)
			if len(handlerPreHandlers) == 0 {
				handlerPreHandlers = pluginPreHandlers
			}
			if len(handlerMidHandlers) == 0 {
				handlerMidHandlers = pluginMidHandlers
			}
			engine.UsePreHandler(append(p.getMiddlewares(handler.Middlewares.PreHandlers), handlerPreHandlers...)...)
			engine.UseMidHandler(append(p.getMiddlewares(handler.Middlewares.MidHandlers), handlerMidHandlers...)...)

			// process triggers
			if len(handler.Triggers.FullMatches) > 0 {
				p.attachMatcher(engine.OnFullMatchGroup(handler.Triggers.FullMatches), handler).Handle(endpoints[handler.HandlerName])
			}
			if len(handler.Triggers.Keywords) > 0 {
				p.attachMatcher(engine.OnKeywordGroup(handler.Triggers.Keywords), handler).Handle(endpoints[handler.HandlerName])
			}
			if len(handler.Triggers.Commands) > 0 {
				p.attachMatcher(engine.OnCommandGroup(handler.Triggers.Commands), handler).Handle(endpoints[handler.HandlerName])
			}
			if len(handler.Triggers.Prefixes) > 0 {
				p.attachMatcher(engine.OnPrefixGroup(handler.Triggers.Prefixes), handler).Handle(endpoints[handler.HandlerName])
			}
			if len(handler.Triggers.Suffixes) > 0 {
				p.attachMatcher(engine.OnSuffixGroup(handler.Triggers.Suffixes), handler).Handle(endpoints[handler.HandlerName])
			}
			for _, regex := range handler.Triggers.Regexes {
				p.attachMatcher(engine.OnRegex(regex), handler).Handle(endpoints[handler.HandlerName])
			}
			if handler.Triggers.Notice {
				p.attachMatcher(engine.OnNotice(), handler).Handle(endpoints[handler.HandlerName])
			}

			// logging initialization
			logging.Info(logger.NewFields().WithMessage("plugin initialized").WithData(map[string]any{"name": plugin.Name, "handler": handler.HandlerName, "metadata": handler}))
		}
	}
}

func (p *pluginManager) parseGroups(groups []PluginGroup) {
	for _, group := range groups {
		p.groups.Set(group.Name, &HandlerMiddlewareMetadata{
			PreHandlers: group.PreHandlers,
			MidHandlers: group.MidHandlers,
		})

		for _, subGroup := range group.SubGroups {
			subGroup.PreHandlers = append(group.PreHandlers, subGroup.PreHandlers...)
			subGroup.MidHandlers = append(group.MidHandlers, subGroup.MidHandlers...)
		}

		p.parseGroups(group.SubGroups)
	}
}

func (p *pluginManager) getGroupPreHandlers(group string) []zero.Rule {
	if group, got := p.groups.Get(group); got {
		var preHandlers []zero.Rule
		for _, preHandler := range group.PreHandlers {
			if middleware, got := p.middlewares.Get(preHandler); got {
				preHandlers = append(preHandlers, middleware)
			}
		}
		return preHandlers
	}

	return nil
}

func (p *pluginManager) getGroupMidHandlers(group string) []zero.Rule {
	if group, got := p.groups.Get(group); got {
		var midHandlers []zero.Rule
		for _, midHandler := range group.MidHandlers {
			if middleware, got := p.middlewares.Get(midHandler); got {
				midHandlers = append(midHandlers, middleware)
			}
		}
		return midHandlers
	}

	return nil
}

func (p *pluginManager) getMiddlewares(middlewares []string) []zero.Rule {
	var result []zero.Rule
	for _, middleware := range middlewares {
		if middleware, got := p.middlewares.Get(middleware); got {
			result = append(result, middleware)
		}
	}

	return result
}

func (p *pluginManager) parseLimiter(limiter string) (need bool, result func(*zero.Ctx) *rate.Limiter) {
	switch limiter {
	case "user":
		return true, ctxext.LimitByUser
	case "group":
		return true, ctxext.LimitByGroup
	default:
		return false, nil
	}
}

func (p *pluginManager) attachMatcher(matcher *control.Matcher, metadata HandlerMetadata) *control.Matcher {
	if limited, limiter := p.parseLimiter(metadata.Limiter); limited {
		return matcher.Limit(limiter).SetBlock(metadata.Blocked)
	}

	return matcher.SetBlock(metadata.Blocked)
}
