package mux

import (
	"forester/internal/tmpl"
	"net/http"

	"github.com/go-chi/chi/v5"
)
import "github.com/go-chi/render"

func MountKickstart(r *chi.Mux) {
	r.Use(render.SetContentType(render.ContentTypePlainText))
	r.Use(DebugMiddleware)
	r.Get("/", HandleKickstart)
	r.Post("/register", HandleRegister)
}

func HandleKickstart(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	err := tmpl.RenderKickstartDiscover(w)
	if err != nil {
		renderGrubError(err, w, r)
		return
	}
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
