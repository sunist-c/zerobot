package openai

import (
	"github.com/FloatTech/ZeroBot-Plugin/cls"
	"github.com/FloatTech/ZeroBot-Plugin/manager"
	"github.com/alioth-center/infrastructure/logger"
	"github.com/alioth-center/infrastructure/thirdparty/openai"
	"github.com/alioth-center/infrastructure/trace"
	"github.com/alioth-center/infrastructure/utils/values"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"time"
)

var (
	client     openai.Client
	cfg        OpenPlatformConfig
	replyCache map[int64]map[int64][]sessionMessage
)

type sessionMessage struct {
	role    openai.ChatRoleEnum
	message string
	created time.Time
}

type OpenPlatformConfig struct {
	Secret    openai.Config      `yaml:"secret"`
	Prompt    string             `yaml:"prompt"`
	Model     string             `yaml:"model"`
	Price     map[string]float64 `yaml:"price"`
	CacheTime float64            `yaml:"cache_time"`
}

func init() {
	manager.Default().RegisterHandler("openai-agent-handler", Agent{})
	manager.InitBeforeServe(func() {
		err := manager.GetYamlPublicConfig(&cfg, "openai_agent_configs")
		if err != nil {
			panic(err)
		}

		replyCache = map[int64]map[int64][]sessionMessage{}
		client = openai.NewClient(cfg.Secret, muteLogger{})
	})
}

type Agent struct {
	manager.BaseHandler
}

func (m Agent) HandleFunc(ctx *zero.Ctx) {
	x := trace.NewContext()

	if ctx.ExtractPlainText() == "" {
		ctx.SendChain(message.Text("小酒在这里，有什么可以帮助你的吗？"))
		return
	}

	// 查询缓存信息
	requestMessage := values.BuildStrings("发送人:", ctx.Event.Sender.NickName, "\n", "发送信息:", ctx.ExtractPlainText())
	cachedMessages := m.searchCachedMessages(ctx)
	cachedMessages = append(cachedMessages, sessionMessage{role: openai.ChatRoleEnumUser, message: requestMessage})
	messages := []openai.ChatMessageObject{{Role: openai.ChatRoleEnumSystem, Content: cfg.Prompt}}
	for _, cachedMessage := range cachedMessages {
		messages = append(messages, openai.ChatMessageObject{
			Role:    cachedMessage.role,
			Content: cachedMessage.message,
		})
	}

	// 请求OpenAI
	res, err := client.CompleteChat(openai.CompleteChatRequest{
		Body: openai.CompleteChatRequestBody{
			Model:     cfg.Model,
			Messages:  messages,
			N:         1,
			MaxTokens: 1500,
		},
	})
	if err != nil {
		cls.Logger().Error(logger.NewFields(x).WithMessage("ai reply failed").WithData(err.Error()))
		ctx.SendChain(message.Text("小酒，坏掉了惹...", err.Error()))
		return
	}
	if len(res.Choices) == 0 {
		cls.Logger().Error(logger.NewFields(x).WithMessage("ai reply failed").WithData("no choices"))
		ctx.SendChain(message.Text("有神秘势力在干扰我...", "没有返回结果"))
		return
	}

	// 记录日志
	cost := (cfg.Price["input"]*float64(res.Usage.PromptTokens) + cfg.Price["output"]*float64(res.Usage.CompletionTokens)) / 1000
	cls.Logger().Info(logger.NewFields(x).
		WithMessage("ai reply generated").
		WithData(map[string]any{"input": ctx.ExtractPlainText(), "output": res.Choices[0].Message.Content}).
		WithField("model", cfg.Model).
		WithField("cost", cost).
		WithField("input_token", res.Usage.PromptTokens).
		WithField("output_token", res.Usage.CompletionTokens).
		WithField("total_token", res.Usage.TotalTokens),
	)

	// 将用户信息记录到缓存
	m.addCacheMessages(ctx, ctx.ExtractPlainText(), res.Choices[0].Message.Content)

	// 返回结果
	ctx.SendChain(
		message.Reply(ctx.Event.MessageID),
		message.Text(res.Choices[0].Message.Content),
	)
}

func (m Agent) AttachRules() []zero.Rule {
	return []zero.Rule{zero.OnlyToMe}
}

func (m Agent) searchCachedMessages(ctx *zero.Ctx) []sessionMessage {
	// 查询信息缓存
	var uid, gid int64
	if ctx.Event.MessageType == "group" {
		uid, gid = ctx.Event.UserID, ctx.Event.GroupID
	} else {
		uid, gid = ctx.Event.UserID, ctx.Event.UserID
	}

	// 不存在则创建
	groupCache, hasGroupCache := replyCache[gid]
	if !hasGroupCache {
		replyCache[gid] = map[int64][]sessionMessage{}
		groupCache = replyCache[gid]
	}
	userCache, hasUserCache := groupCache[uid]
	if !hasUserCache {
		replyCache[gid][uid] = []sessionMessage{}
		userCache = replyCache[gid][uid]
	}

	// 查找未过期的信息，过期则清除
	var result []sessionMessage
	for i, sessionMessage := range userCache {
		if time.Since(sessionMessage.created).Seconds() < cfg.CacheTime && sessionMessage.role == openai.ChatRoleEnumUser {
			result = append(result, userCache[i], userCache[i+1])
		}
	}

	replyCache[gid][uid] = result
	return result
}

func (m Agent) addCacheMessages(ctx *zero.Ctx, input, output string) {
	// 查询信息缓存
	var uid, gid int64
	if ctx.Event.MessageType == "group" {
		uid, gid = ctx.Event.UserID, ctx.Event.GroupID
	} else {
		uid, gid = ctx.Event.UserID, ctx.Event.UserID
	}

	// 不存在则创建
	groupCache, hasGroupCache := replyCache[gid]
	if !hasGroupCache {
		replyCache[gid] = map[int64][]sessionMessage{}
		groupCache = replyCache[gid]
	}
	userCache, hasUserCache := groupCache[uid]
	if !hasUserCache {
		replyCache[gid][uid] = []sessionMessage{}
		userCache = replyCache[gid][uid]
	}

	// 添加信息
	userCache = append(userCache, sessionMessage{role: openai.ChatRoleEnumUser, message: input, created: time.Now()})
	userCache = append(userCache, sessionMessage{role: openai.ChatRoleEnumAssistant, message: output, created: time.Now()})
	replyCache[gid][uid] = userCache
}
