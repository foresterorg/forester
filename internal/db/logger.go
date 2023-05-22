package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/tracelog"
	"golang.org/x/exp/slog"
)

// Logger implements the tracelog.Logger interface by wrapping a slog.Logger
// https://godocs.io/github.com/jackc/pgx/v5/tracelog
// https://godocs.io/github.com/jackc/pgx/v5#QueryTracer
// https://godocs.io/golang.org/x/exp/slog
type Logger struct {
	slogger *slog.Logger
}

func NewTracerLogger(l *slog.Logger, level tracelog.LogLevel) pgx.QueryTracer {
	return &tracelog.TraceLog{
		Logger:   &Logger{slogger: l},
		LogLevel: level,
	}
}

func (l *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	var attrs []slog.Attr
	for k, v := range data {
		attrs = append(attrs, slog.Any(k, v))
	}
	l.slogger.LogAttrs(ctx, translateLevel(level), msg, attrs...)
}

func translateLevel(level tracelog.LogLevel) slog.Level {
	switch level {
	case tracelog.LogLevelTrace:
		return slog.LevelDebug
	case tracelog.LogLevelDebug:
		return slog.LevelDebug
	case tracelog.LogLevelInfo:
		return slog.LevelInfo
	case tracelog.LogLevelWarn:
		return slog.LevelWarn
	case tracelog.LogLevelError:
		return slog.LevelError
	case tracelog.LogLevelNone:
		return slog.LevelError
	default:
		return slog.LevelError
	}
}
