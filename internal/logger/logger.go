package logger

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	LogFormatText = "text"
	LogFormatJson = "json"
)

const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)

func New(format string, level string) *slog.Logger {
	var loggerLevel slog.Level
	var unknownLevel bool
	switch level {
	case LevelDebug:
		loggerLevel = slog.LevelDebug
	case LevelInfo:
		loggerLevel = slog.LevelInfo
	case LevelWarn:
		loggerLevel = slog.LevelWarn
	case LevelError:
		loggerLevel = slog.LevelError
	default:
		unknownLevel = true
	}

	var logger *slog.Logger
	switch format {
	case LogFormatJson:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel}))
	case LogFormatText:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel}))
	default:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel}))
		logger.Warn(fmt.Sprintf("unsupported logging format %s, using default format instead", format))
	}

	if unknownLevel {
		logger.Warn(fmt.Sprintf("unsupported logging level %s, using default info level instead", level))
	}

	return logger
}
