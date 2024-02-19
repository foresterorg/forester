package mux

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"

	"forester/internal/config"
)

func MountLogs(r *chi.Mux) {
	r.Head("/*", serveLogsPath)
	r.Get("/*", serveLogsPath)
}

func serveLogsPath(w http.ResponseWriter, r *http.Request) {
	fs := http.StripPrefix("/logs", http.FileServer(http.Dir(config.Logging.SyslogDir)))
	fs.ServeHTTP(w, r)
}
