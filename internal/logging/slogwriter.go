package logging

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/exp/slog"
)

// SlogWriter writes Go standard library logger to slog as well as
// to standard error. It is used to prevent "missed messages"
// of some libraries which can possibly write via stdlib log.
type SlogWriter struct {
	Logger *slog.Logger
	Level  slog.Level
}

func (slw SlogWriter) Write(p []byte) (n int, err error) {
	slw.Logger.Log(context.Background(), slw.Level, string(p))
	fmt.Fprintln(os.Stderr, string(p))

	return len(p), nil
}
