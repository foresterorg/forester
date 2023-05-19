package log

import (
	"context"

	"golang.org/x/exp/slog"
)

// A ContextHandler adds fields from context.
type ContextHandler struct {
	handler slog.Handler
}

func NewContextHandler(h slog.Handler) *ContextHandler {
	// Optimization: avoid chains of ContextHandlers.
	if lh, ok := h.(*ContextHandler); ok {
		h = lh.Handler()
	}
	return &ContextHandler{h}
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewContextHandler(h.handler.WithAttrs(attrs))
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return NewContextHandler(h.handler.WithGroup(name))
}

func (h *ContextHandler) Handler() slog.Handler {
	return h.handler
}
