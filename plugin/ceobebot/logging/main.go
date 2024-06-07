package logging

import (
	"encoding/csv"
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/cls"
	"github.com/FloatTech/ZeroBot-Plugin/manager"
	"github.com/alioth-center/infrastructure/exit"
	"github.com/alioth-center/infrastructure/logger"
	"github.com/alioth-center/infrastructure/utils/values"
	zero "github.com/wdvxdr1123/ZeroBot"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var (
	messageBuffer = make(chan *manager.ReceivedMessage, 100)
	syncMutex     = sync.Mutex{}
	signalIndex   = 0
)

func init() {
	manager.Default().RegisterHandler("message-logging-collector-handler", MessageCollector{})
	manager.Default().RegisterMiddleware("message-logging-collector-middleware", manager.LogMessage())
	go bufferLog()
}

type MessageCollector struct {
	manager.BaseHandler
}

func (m MessageCollector) HandleFunc(ctx *zero.Ctx) {
	message := &manager.ReceivedMessage{
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
		message.Text = "[media message]"
	}

	if len(messageBuffer) > cap(messageBuffer)/2 {
		signalIndex++
		writeBuffer(signalIndex)
	}

	syncMutex.Lock()
	messageBuffer <- message
	syncMutex.Unlock()
}

func bufferLog() {
	// write buffer to file immediately when the program exits
	exit.Register(func(sig string) string {
		writeBuffer(-1)
		return "message-logging-collector-handler is exiting"
	}, "message-logging-collector-handler")

	for {
		select {
		case <-time.After(time.Hour):
			writeBuffer(0)
		}
	}
}

func writeBuffer(signal int) {
	syncMutex.Lock()
	defer syncMutex.Unlock()

	if len(messageBuffer) == 0 {
		return
	}

	messageArray, complete := make([]*manager.ReceivedMessage, 0, len(messageBuffer)), false
	for {
		select {
		case m := <-messageBuffer:
			messageArray = append(messageArray, m)
		default:
			complete = true
			break
		}

		if complete {
			break
		}
	}

	suffix := ""
	switch signal {
	case -1:
		suffix = "_exit"
	case 0:
		suffix = "_hourly"
	default:
		suffix = "_flush" + strconv.Itoa(signal)
	}

	filename := filepath.Join("./data/message_partition", values.BuildStrings(time.Now().Format("2006_01_02_15"), suffix, ".csv"))
	messageFile, openErr := os.Create(filename)
	if openErr != nil {
		cls.Logger().Error(logger.NewFields().WithMessage("failed to create message partition file").WithData(openErr.Error()).WithField("payload", &messageArray))
		return
	}

	writer := csv.NewWriter(messageFile)

	// 写入 CSV 表头
	header := []string{"ID", "Text", "Type", "Name", "Group"}
	if err := writer.Write(header); err != nil {
		cls.Logger().Error(logger.NewFields().WithMessage("failed to write csv header").WithData(err.Error()).WithField("payload", &messageArray))
		return
	}

	// 写入 CSV 内容
	for _, msg := range messageArray {
		record := []string{
			strconv.FormatInt(msg.ID, 10),
			msg.Text,
			msg.Type,
			msg.Name,
			strconv.FormatInt(msg.Group, 10),
		}
		if err := writer.Write(record); err != nil {
			cls.Logger().Error(logger.NewFields().WithMessage("failed to write csv content").WithData(err.Error()).WithField("payload", &messageArray))
			return
		}
	}

	writer.Flush()

	closeErr := messageFile.Close()
	if closeErr != nil {
		cls.Logger().Error(logger.NewFields().WithMessage("failed to close message partition file").WithData(closeErr.Error()))
	}

	fmt.Println("message partition file written:", filename)
	cls.Logger().Info(logger.NewFields().WithMessage("message partition file written").WithField("filename", filename))
}
