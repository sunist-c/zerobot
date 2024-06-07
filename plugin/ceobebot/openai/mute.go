package openai

import "github.com/alioth-center/infrastructure/logger"

type muteLogger struct{}

func (m muteLogger) Debug(fields logger.Fields) {}

func (m muteLogger) Info(fields logger.Fields) {}

func (m muteLogger) Warn(fields logger.Fields) {}

func (m muteLogger) Error(fields logger.Fields) {}

func (m muteLogger) Fatal(fields logger.Fields) {}

func (m muteLogger) Panic(fields logger.Fields) {}

func (m muteLogger) Log(level logger.Level, fields logger.Fields) {}

func (m muteLogger) Logf(level logger.Level, fields logger.Fields, format string, args ...any) {}

func (m muteLogger) Debugf(fields logger.Fields, format string, args ...any) {}

func (m muteLogger) Infof(fields logger.Fields, format string, args ...any) {}

func (m muteLogger) Warnf(fields logger.Fields, format string, args ...any) {}

func (m muteLogger) Errorf(fields logger.Fields, format string, args ...any) {}

func (m muteLogger) Fatalf(fields logger.Fields, format string, args ...any) {}

func (m muteLogger) Panicf(fields logger.Fields, format string, args ...any) {}
