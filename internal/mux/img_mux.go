package mux

import (
	"context"
	"forester/internal/config"
	"forester/internal/db"
	"forester/internal/img"
	"forester/internal/model"
	"net/http"
	"strconv"
	"time"

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
		slog.ErrorContext(r.Context(), "invalid ID", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dao := db.GetImageDao(r.Context())
	dbImage, err := dao.FindByID(r.Context(), id)
	if err != nil {
		slog.ErrorContext(r.Context(), "cannot find image with this id", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	n, err := img.Copy(r.Context(), id, r.Body)
	if err != nil {
		slog.ErrorContext(r.Context(), "cannot copy image", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.DebugContext(r.Context(), "image written", "size", n)

	go extractImage(dbImage)
}

func extractImage(dbImage *model.Image) {
	deadline := time.Now().Add(30 * time.Minute)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	result, err := img.Extract(ctx, dbImage.ID)
	if err != nil {
		slog.ErrorContext(ctx, "error during extraction", "err", err)
		return
	}

	dbImage.LiveimgSha256 = result.LiveimgSha256
	dao := db.GetImageDao(ctx)
	err = dao.Update(ctx, dbImage)
	if err != nil {
		slog.ErrorContext(ctx, "could not update image sha256", "err", err)
	}
}

func serveImagePath(w http.ResponseWriter, r *http.Request) {
	fs := http.StripPrefix("/img", http.FileServer(http.Dir(config.Images.Directory)))
	fs.ServeHTTP(w, r)
}
