package openai

import (
	"github.com/FloatTech/ZeroBot-Plugin/cls"
	"github.com/FloatTech/ZeroBot-Plugin/manager"
	"github.com/alioth-center/infrastructure/logger"
	"github.com/alioth-center/infrastructure/thirdparty/openai"
	"github.com/alioth-center/infrastructure/trace"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	client openai.Client
	cfg    PublicConfig
)

type PublicConfig struct {
	Secret openai.Config      `yaml:"secret"`
	Prompt string             `yaml:"prompt"`
	Model  string             `yaml:"model"`
	Price  map[string]float64 `yaml:"price"`
}

func lazyLoad() {
	if client == nil {
		err := manager.GetYamlPublicConfig(&cfg, "openai_agent_configs")
		if err != nil {
			panic(err)
		}

		client = openai.NewClient(cfg.Secret, muteLogger{})
	}
}

func init() {
	manager.Default().RegisterHandler("openai-agent-handler", Agent{})
}

type Agent struct {
	manager.BaseHandler
}

func (m Agent) HandleFunc(ctx *zero.Ctx) {
	lazyLoad()
	x := trace.NewContext()

	if ctx.ExtractPlainText() == "" {
		ctx.SendChain(message.Text("小酒在这里，有什么可以帮助你的吗？"))
		return
	}

	res, err := client.CompleteChat(openai.CompleteChatRequest{
		Body: openai.CompleteChatRequestBody{
			Model:     cfg.Model,
			Messages:  []openai.ChatMessageObject{{Role: "system", Content: cfg.Prompt}, {Role: "user", Content: ctx.ExtractPlainText()}},
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

	ctx.SendChain(message.Text(res.Choices[0].Message.Content))
}

func (m Agent) AttachRules() []zero.Rule {
	return []zero.Rule{zero.OnlyToMe}
}
