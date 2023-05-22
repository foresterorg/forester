package logging

import (
	"context"

	"golang.org/x/exp/slog"
)

type SlogWriter struct {
	Logger *slog.Logger
	Level  slog.Level
}

func (slw SlogWriter) Write(p []byte) (n int, err error) {
	slw.Logger.Log(context.Background(), slw.Level, string(p))

	return len(p), nil
}
