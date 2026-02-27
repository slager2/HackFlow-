package logger

import (
	"log/slog"
	"os"
)

// Setup initializes the global structured logger.
// In production, it uses JSON format; in development, text format.
func Setup(env string) {
	var handler slog.Handler

	if env == "production" || env == "prod" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
