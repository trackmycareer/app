package logger

import (
	"log/slog"
	"os"
)

// Init configures the default slog logger based on the environment.
// In production it outputs structured JSON at Info level; otherwise it
// uses human-readable text at Debug level.
func Init(env string) {
	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	slog.SetDefault(slog.New(handler))
}
