package mux

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)
import "github.com/go-chi/render"

func MountBoot(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypePlainText))
		r.Get("/shim.efi", HandleShim)
	})
}

func HandleShim(w http.ResponseWriter, r *http.Request) {
	// ServeFile instead - parametrized
}
