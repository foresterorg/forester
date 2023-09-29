package mux

import (
	"errors"
	"fmt"
	"forester/internal/db"
	"forester/internal/model"
	"forester/internal/tmpl"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/slog"
)
import "github.com/go-chi/render"

func MountKickstart(r *chi.Mux) {
	r.Use(render.SetContentType(render.ContentTypePlainText))
	r.Use(DebugMiddleware)
	r.Get("/", HandleKickstart)
	r.Post("/register", HandleRegister)
}

var ErrMACHeaderInvalid = errors.New("invalid format of RHN MAC header")

func HandleKickstart(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	var system *model.System
	var err error

	sDao := db.GetSystemDao(r.Context())
	for i := 0; i < 64; i++ {
		if macString := r.Header.Get(fmt.Sprintf("X-RHN-Provisioning-MAC-%d", i)); macString != "" {
			headerLine := strings.SplitN(macString, " ", 2)
			if len(headerLine) != 2 {
				renderKsError(ErrMACHeaderInvalid, w, r)
				return
			}
			macString = headerLine[1]
			slog.DebugContext(r.Context(), "searching for system", "mac", macString)
			mac, err := net.ParseMAC(macString)
			if err != nil {
				renderKsError(err, w, r)
				return
			}

			system, err = sDao.FindByMac(r.Context(), mac)
			if errors.Is(err, pgx.ErrNoRows) {
				slog.DebugContext(r.Context(), "unknown system", "mac", macString)
			} else if err != nil {
				slog.ErrorContext(r.Context(), "error while finding system", "mac", macString, "err", err)
				renderKsError(err, w, r)
				return
			} else {
				break
			}
		} else {
			break
		}
	}

	if system != nil && system.Installable() {
		err = tmpl.RenderKickstartInstall(w, tmpl.KickstartParams{ImageID: *system.ImageID})
	} else if system != nil && !system.Installable() {
		slog.WarnContext(r.Context(), "system found but not installable",
			"id", system.ID,
			"name", system.Name,
			"acquired_at", system.AcquiredAt.String(),
			"image_id", system.ImageID)
		err = tmpl.RenderKickstartDiscover(w)
	} else {
		err = tmpl.RenderKickstartDiscover(w)
	}
	if err != nil {
		renderKsError(err, w, r)
		return
	}
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func renderKsError(ksErr error, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	slog.ErrorContext(r.Context(), "rendering error as kickstart comment", "err", ksErr)
	err := tmpl.RenderKickstartError(w, tmpl.KickstartErrorParams{Message: ksErr.Error()})
	if err != nil {
		slog.ErrorContext(r.Context(), "cannot render template", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
