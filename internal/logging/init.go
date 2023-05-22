package logging

import (
	"log"
	"os"

	"golang.org/x/exp/slog"
)

var emptyAttr = slog.Attr{}

// Initialize configures logging system. Use slog package to create log entries.
// Make sure to use context variants when context is available, for example
// slog.InfoCtx.
func Initialize(level slog.Level) {
	th := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return emptyAttr
			}
			if a.Key == slog.LevelKey {
				return emptyAttr
			}
			return a
		},
	})
	logger := slog.New(NewContextHandler(th))
	slog.SetDefault(logger)

	log.SetOutput(SlogWriter{logger, slog.LevelInfo})
}
