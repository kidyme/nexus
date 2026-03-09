package log

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// Init 初始化默认 logger：生产用 JSON，开发用 tint 彩色输出。
func Init(isProd bool) {
	var logger *slog.Logger
	if isProd {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		}))
	} else {
		logger = slog.New(tint.NewHandler(os.Stdout, &tint.Options{
			Level:     slog.LevelDebug,
			AddSource: true,
		}))
	}
	slog.SetDefault(logger)
}

// Logger 返回默认 logger，便于在少数场景下使用原生 slog 能力。
func Logger() *slog.Logger {
	return slog.Default()
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
}
