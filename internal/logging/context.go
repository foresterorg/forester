package logging

import (
	"context"

	"github.com/thanhpk/randstr"
)

type ctxKeyId int

const (
	traceCtxKey ctxKeyId = iota
)

// TraceId returns request id or an empty string when not set.
func TraceId(ctx context.Context) string {
	value := ctx.Value(traceCtxKey)
	if value == nil {
		return ""
	}
	return value.(string)
}

// WithTraceId returns context copy with trace id value.
func WithTraceId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceCtxKey, id)
}

func NewTraceId() string {
	return randstr.Base62(8)
}
