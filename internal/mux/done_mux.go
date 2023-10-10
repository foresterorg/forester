package mux

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
)

func MountDone(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypePlainText))

		r.Post("/{ID}", HandleDone)
	})
}

func HandleDone(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "ID"), 10, 64)
	if err != nil {
		slog.InfoContext(r.Context(), "installation what", "system_id", chi.URLParam(r, "ID"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slog.InfoContext(r.Context(), "installation done", "system_id", id)
	// TODO: configure booting from HDD in case of libvirt and power it on

	w.WriteHeader(http.StatusOK)
}
