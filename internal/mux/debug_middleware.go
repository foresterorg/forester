package mux

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
)

func DebugMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var headers []slog.Attr
		for k, v := range r.Header {
			pair := slog.String(k, strings.Join(v, " "))
			headers = append(headers, pair)
		}
		if len(headers) > 0 {
			slog.DebugContext(r.Context(), "request headers", "headers", slog.GroupValue(headers...))
		}

		sb := strings.Builder{}
		wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		wrw.Tee(&sb)

		next.ServeHTTP(wrw, r)
		if slog.Default().Enabled(r.Context(), slog.LevelDebug) {
			slog.DebugContext(r.Context(), "payload contents", "payload", sb.String())
		}
	}
	return http.HandlerFunc(fn)
}
