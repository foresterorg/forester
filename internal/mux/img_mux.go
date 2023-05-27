package mux

import (
	"forester/internal/config"
	"forester/internal/img"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
)

func MountImages(r *chi.Mux) {
	r.Head("/*", serveImagePath)
	r.Get("/*", serveImagePath)
	r.Put("/{ID}", uploadImage)
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	if !HasContentType(r, "application/octet-stream") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "ID"), 10, 64)
	if err != nil {
		slog.ErrorCtx(r.Context(), "invalid ID", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	n, err := img.Copy(r.Context(), id, r.Body)
	if err != nil {
		slog.ErrorCtx(r.Context(), "cannot copy image", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.DebugCtx(r.Context(), "image written", "size", n)

	go img.Extract(r.Context(), id)
}

func serveImagePath(w http.ResponseWriter, r *http.Request) {
	fs := http.StripPrefix("/img", http.FileServer(http.Dir(config.Images.Directory)))
	fs.ServeHTTP(w, r)
}
