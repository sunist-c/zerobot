package logging

import (
	"github.com/FloatTech/ZeroBot-Plugin/manager"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	manager.Default().RegisterHandler("message-logging-collector-handler", MessageCollector{})
	manager.Default().RegisterMiddleware("message-logging-collector-middleware", manager.LogMessage())
}

type MessageCollector struct {
	manager.BaseHandler
}

func (m MessageCollector) HandleFunc(_ *zero.Ctx) {}
