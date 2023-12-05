package mux

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"forester/internal/config"
	"forester/internal/db"
	"forester/internal/img"
	"forester/internal/logging"
	"forester/internal/model"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	err = ensureDir(dbImage.ID)
	if err != nil {
		slog.ErrorContext(r.Context(), "cannot create dir for ISO", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	n, isoSha256, err := img.Copy(r.Context(), isoPath(dbImage.ID), r.Body)
	if err != nil {
		slog.ErrorContext(r.Context(), "cannot copy image", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	slog.DebugContext(r.Context(), "image written", "size", n, "sha256sum", isoSha256)

	dbImage.IsoSha256 = isoSha256
	err = dao.Update(r.Context(), dbImage)
	if err != nil {
		slog.ErrorContext(r.Context(), "could not update ISO sha256", "err", err)
	}

	go extractImage(dbImage)
}

func ensureDir(imageId int64) error {
	result := filepath.Join(config.Images.Directory, strconv.FormatInt(imageId, 10))
	err := os.MkdirAll(result, 0744)

	if err != nil {
		return fmt.Errorf("cannot write image: %w", err)
	}

	return nil
}

func dirPath(imageId int64) string {
	return filepath.Join(config.Images.Directory, strconv.FormatInt(imageId, 10))
}

func isoPath(imageId int64) string {
	return filepath.Join(config.Images.Directory, strconv.FormatInt(imageId, 10), "image.iso")
}

func sha256sum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("cannot open %s: %w", file, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func fileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func extractImage(dbImage *model.Image) {
	deadline := time.Now().Add(30 * time.Minute)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	ctx = logging.WithJobId(ctx, logging.NewJobId())
	imagePath := dirPath(dbImage.ID)

	err := ensureDir(dbImage.ID)
	if err != nil {
		slog.ErrorContext(ctx, "error during extraction", "err", err)
		return
	}

	slog.DebugContext(ctx, "extracting ISO image", "img", imagePath)
	err = img.ExtractToDir(ctx, isoPath(dbImage.ID), imagePath)
	if err != nil {
		slog.ErrorContext(ctx, "error during extraction", "err", err)
		return
	}

	slog.DebugContext(ctx, "generating boot.iso image", "img", imagePath)
	err = img.GenerateBootISO(ctx, dbImage.ID, imagePath)
	if err != nil {
		slog.ErrorContext(ctx, "error during boot.iso generation", "err", err)
		return
	}

	err = os.Symlink("./EFI/BOOT/BOOTX64.EFI", filepath.Join(imagePath, "shim.efi"))
	if err != nil {
		slog.ErrorContext(ctx, "cannot create symlink", "err", err)
		return
	}
	err = os.Symlink("./EFI/BOOT/grubx64.efi", filepath.Join(imagePath, "grubx64.efi"))
	if err != nil {
		slog.ErrorContext(ctx, "cannot create symlink", "err", err)
		return
	}

	imgPath := filepath.Join(imagePath, "liveimg.tar.gz")
	slog.DebugContext(ctx, "checking for liveimg.tar.gz", "path", imgPath)
	if ok, err := fileExists(imgPath); ok && (err == nil) {
		sum, err := sha256sum(imgPath)
		if err != nil {
			slog.ErrorContext(ctx, "error calculate SHA256 sum", "err", err)
			return
		}
		slog.DebugContext(ctx, "sha256 of the liveimg.tar.gz", "sha256sum", sum)

		dbImage.LiveimgSha256 = sum
		dao := db.GetImageDao(ctx)
		err = dao.Update(ctx, dbImage)
		if err != nil {
			slog.ErrorContext(ctx, "could not update image sha256", "err", err)
		}
	}
}

func serveImagePath(w http.ResponseWriter, r *http.Request) {
	fs := http.StripPrefix("/img", http.FileServer(http.Dir(config.Images.Directory)))
	fs.ServeHTTP(w, r)
}
