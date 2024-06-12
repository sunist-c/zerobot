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
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type ProphecyConfig struct {
	Results []ProphecyItem     `yaml:"results"`
	Prompt  string             `yaml:"prompt"`
	Model   string             `yaml:"model"`
	Price   map[string]float64 `yaml:"price"`
}

type ProphecyItem struct {
	Summary    string `yaml:"summary"`
	Lucky      int    `yaml:"lucky"`
	Text       string `yaml:"text"`
	Annotation string `yaml:"annotation"`
}

var (
	prophecyCfg  ProphecyConfig
	sortedResult map[int]ProphecyItem
)

func init() {
	manager.Default().RegisterHandler("prophecy-agent-handler", ProphecyAgent{})
	manager.InitBeforeServe(func() {
		if len(prophecyCfg.Results) == 0 {
			err := manager.GetYamlPublicConfig(&prophecyCfg, "prophecy_configs")
			if err != nil {
				panic(err)
			}
		}

		sortedResult = map[int]ProphecyItem{}
		for _, result := range prophecyCfg.Results {
			sortedResult[result.Lucky] = result
		}
	})
}

type ProphecyAgent struct {
	manager.BaseHandler
}

func (m ProphecyAgent) HandleFunc(ctx *zero.Ctx) {
	x := trace.NewContext()
	index := m.generateIndex(ctx.Event.UserID)
	summary, text, annotation, lucky := sortedResult[index].Summary, sortedResult[index].Text, sortedResult[index].Annotation, sortedResult[index].Lucky
	requestMessage := values.BuildStrings("签文:", text, "解签:", summary, "注解:", annotation, "幸运值:", strconv.Itoa(lucky))
	res, err := client.CompleteChat(openai.CompleteChatRequest{
		Body: openai.CompleteChatRequestBody{
			Model:     prophecyCfg.Model,
			Messages:  []openai.ChatMessageObject{{Role: "system", Content: prophecyCfg.Prompt}, {Role: "user", Content: requestMessage}},
			N:         1,
			MaxTokens: 1000,
		},
	})

	if err != nil {
		cls.Logger().Error(logger.NewFields(x).WithMessage("prophecy generate failed").WithData(err.Error()))
		ctx.SendChain(message.Text("签文解析失败...今天也许不宜求签呜~", err.Error()))
		return
	}
	if len(res.Choices) == 0 {
		cls.Logger().Error(logger.NewFields(x).WithMessage("prophecy generate failed").WithData("no choices"))
		ctx.SendChain(message.Text("签文空空如也...也许你今天的运势被天机所扰喵！", "没有返回结果"))
		return
	}

	cost := (prophecyCfg.Price["input"]*float64(res.Usage.PromptTokens) + prophecyCfg.Price["output"]*float64(res.Usage.CompletionTokens)) / 1000
	cls.Logger().Info(logger.NewFields(x).
		WithMessage("prophecy generated").
		WithData(sortedResult[index]).
		WithField("model", prophecyCfg.Model).
		WithField("cost", cost).
		WithField("input_token", res.Usage.PromptTokens).
		WithField("output_token", res.Usage.CompletionTokens).
		WithField("total_token", res.Usage.TotalTokens),
	)

	ctx.SendChain(message.Text(values.BuildStringsWithJoin("\n",
		"运势："+strings.Repeat("★", lucky+1),
		"签文："+text,
		"解签："+summary,
		res.Choices[0].Message.Content),
	))
}

func (m ProphecyAgent) generateIndex(uid int64) int {
	now := time.Now()
	return rand.New(rand.NewSource(uid + int64(now.Year()*1000+now.YearDay())<<now.Month())).Intn(8)
}

func (m ProphecyAgent) AttachRules() []zero.Rule {
	return []zero.Rule{zero.OnlyToMe}
}
