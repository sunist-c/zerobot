package manager

import (
	"github.com/alioth-center/infrastructure/logger"
	"github.com/alioth-center/infrastructure/trace"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type ReceivedMessage struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Text string `json:"text"`
}

func LogMessage(log logger.Logger) func(ctx *zero.Ctx) bool {
	if log == nil {
		log = logging
	}

	return func(ctx *zero.Ctx) bool {
		// just log the message to custom logger like loki/logrus/cls
		log.Info(logger.NewFields().WithMessage("message received").WithData(&ReceivedMessage{
			ID:   ctx.Event.UserID,
			Name: ctx.CardOrNickName(ctx.Event.UserID),
			Text: ctx.ExtractPlainText(),
		}))

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
