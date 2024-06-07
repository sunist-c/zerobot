package cls

import (
	"github.com/FloatTech/ZeroBot-Plugin/manager"
	"github.com/alioth-center/infrastructure/logger"
	"github.com/alioth-center/infrastructure/thirdparty/tencent/cls"
)

var logging logger.Logger

func Logger() logger.Logger {
	if logging == nil {
		cfg := cls.Config{}
		err := manager.GetYamlPublicConfig(&cfg, "message_collector_configs")
		if err != nil {
			panic(err)
		}

		fallback := logger.Default()
		clsLogger, err := cls.NewClsLogger(cfg, fallback)
		if err != nil {
			panic(err)
		}

		logging = clsLogger
	}

	return logging
}
