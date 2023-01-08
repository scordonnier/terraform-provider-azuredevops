package logger

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"os"
	"sync"
)

var lock = &sync.Mutex{}
var logger *hclog.Logger

func Debug(_ context.Context, message string) {
	defaultLogger().Debug(message)
}

func Error(_ context.Context, message string) {
	defaultLogger().Error(message)
}

func Info(_ context.Context, message string) {
	defaultLogger().Info(message)
}

func Trace(_ context.Context, message string) {
	defaultLogger().Trace(message)
}

func Warn(_ context.Context, message string) {
	defaultLogger().Warn(message)
}

// Private Methods

func defaultLogger() hclog.Logger {
	if logger == nil {
		defer lock.Unlock()
		lock.Lock()
		level := os.Getenv("TF_LOG_AZDO")
		if level == "" {
			level = "OFF"
		}
		instance := hclog.New(&hclog.LoggerOptions{
			Level:  hclog.LevelFromString(level),
			Output: os.Stdin,
		})
		logger = &instance
	}
	return *logger
}
