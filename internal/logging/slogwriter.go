package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// SlogDualWriter writes Go standard library logger to slog as well as
// to standard error. It is used to prevent "missed messages"
// of some libraries which can possibly write via stdlib log.
type SlogDualWriter struct {
	Logger  *slog.Logger
	Level   slog.Level
	Context context.Context
}

func (slw SlogDualWriter) Write(p []byte) (n int, err error) {
	slw.Logger.Log(slw.Context, slw.Level, strings.TrimSpace(string(p)))
	fmt.Fprintln(os.Stderr, string(p))

	return len(p), nil
}

// SlogWriter writes Go standard library logger to slog.
type SlogWriter struct {
	Logger  *slog.Logger
	Level   slog.Level
	Context context.Context
}

func (slw SlogWriter) Write(p []byte) (n int, err error) {
	slw.Logger.Log(slw.Context, slw.Level, strings.TrimSpace(string(p)))

	return len(p), nil
}
