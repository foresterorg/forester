package mux

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/slog"
)

func DebugMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		sb := strings.Builder{}
		wrw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		wrw.Tee(&sb)

		next.ServeHTTP(wrw, r)
		if slog.Default().Enabled(r.Context(), slog.LevelDebug) {
			slog.DebugCtx(r.Context(), "payload contents", "payload", sb.String())
		}
	}
	return http.HandlerFunc(fn)
}
