package img

import (
	"context"

	"github.com/thanhpk/randstr"
)

type ctxKeyId int

const (
	jobCtxKey ctxKeyId = iota
)

// JobId returns request id or an empty string when not set.
func JobId(ctx context.Context) string {
	value := ctx.Value(jobCtxKey)
	if value == nil {
		return ""
	}
	return value.(string)
}

// WithJobId returns context copy with trace id value.
func WithJobId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, jobCtxKey, id)
}

func NewJobId() string {
	return randstr.Base62(8)
}
