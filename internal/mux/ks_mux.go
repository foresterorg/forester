package mux

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)
import "github.com/go-chi/render"

func MountKickstart(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypePlainText))

		r.Route("/xxxx", func(r chi.Router) {
			r.Get("/", HandleKickstart)
		})

	})
}

func HandleKickstart(w http.ResponseWriter, r *http.Request) {

}
