package mux

import (
	"forester/internal/logging"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
)

func TraceIdMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Edge request id
		tid := r.Header.Get("X-Rh-Edge-Request-Id")
		if tid == "" {
			tid = logging.NewTraceId()
		}
		ctx = logging.WithTraceId(ctx, tid)

		// Store in response headers for easier debugging
		w.Header().Set("X-Trace-Id", tid)

		ctx = logging.WithTraceId(ctx, tid)
		slog.InfoContext(ctx, "started request",
			"method", r.Method,
			"path", r.RequestURI,
			"content_length", r.ContentLength,
		)
		t1 := time.Now()
		next.ServeHTTP(wrw, r.WithContext(ctx))
		slog.InfoContext(ctx, "finished request",
			"method", r.Method,
			"path", r.RequestURI,
			"duration_ms", time.Since(t1).Round(time.Millisecond).String(),
			"status", wrw.Status(),
			"bytes", wrw.BytesWritten(),
		)
	}
	return http.HandlerFunc(fn)
}
