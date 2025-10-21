package logger

import (
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

func Init() *lumberjack.Logger {
	logFilePath := filepath.Join(os.TempDir(), "clispot.log")

	rotator := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    5,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	handler := slog.NewJSONHandler(rotator, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return rotator
}
