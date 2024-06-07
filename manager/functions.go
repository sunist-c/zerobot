package manager

import (
	"github.com/alioth-center/infrastructure/logger"
	"github.com/alioth-center/infrastructure/trace"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type ReceivedMessage struct {
	ID    int64  `json:"id"`
	Text  string `json:"text"`
	Type  string `json:"type"`
	Name  string `json:"name,omitempty"`
	Group int64  `json:"group,omitempty"`
}

func LogMessage() func(ctx *zero.Ctx) bool {
	return func(ctx *zero.Ctx) bool {
		// just log the message to custom logger like loki/logrus/cls
		message := &ReceivedMessage{
			ID:   ctx.Event.UserID,
			Text: ctx.ExtractPlainText(),
			Type: ctx.Event.MessageType,
		}
		switch message.Type {
		case "group":
			message.Group = ctx.Event.GroupID
			message.Name = ctx.CardOrNickName(ctx.Event.UserID)
		case "private":
			message.Group = ctx.Event.UserID
			message.Name = ctx.Event.Sender.Name()
		}

		if message.Text == "" {
			message.Text = ctx.Event.Message.String()
		}

		logging.Info(logger.NewFields().WithMessage("message received").WithData(message))
		return true
	}
}

func pluginEnableLoggingCallback(plugin PluginMetadata) func(_ *zero.Ctx) {
	return func(_ *zero.Ctx) {
		logging.Info(logger.NewFields(trace.NewContext()).WithMessage("plugin enabled").WithData(plugin))
	}
}

func pluginDisableLoggingCallback(plugin PluginMetadata) func(_ *zero.Ctx) {
	return func(_ *zero.Ctx) {
		logging.Info(logger.NewFields(trace.NewContext()).WithMessage("plugin disabled").WithData(plugin))
	}
}
