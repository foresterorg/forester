package mux

import (
	"forester/internal/config"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func MountImages(r *chi.Mux) {
	r.Head("/*", serveImagePath)
	r.Get("/*", serveImagePath)
}

func serveImagePath(w http.ResponseWriter, r *http.Request) {
	fs := http.StripPrefix("/img", http.FileServer(http.Dir(config.Images.Directory)))
	fs.ServeHTTP(w, r)
}
