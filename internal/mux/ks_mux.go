package mux

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/db"
	"forester/internal/model"
	"forester/internal/tmpl"
	"io"
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

func RenderKickstartForSystem(ctx context.Context, system *model.System, w io.Writer) error {
	if system != nil && system.Installable() {
		la := tmpl.RebootLastAction
		aDao := db.GetApplianceDao(ctx)
		sDao := db.GetSnippetDao(ctx)

		appliance, err := aDao.FindByID(ctx, *system.ApplianceID)
		if errors.Is(err, pgx.ErrNoRows) {
			slog.DebugContext(ctx, "installing a system without appliance")
		} else if err != nil {
			slog.ErrorContext(ctx, "error while fetching appliance for system", "id", system.ID)
			return err
		}

		// libvirt cannot be restarted due to boot order hook
		if appliance != nil && appliance.Kind == model.LibvirtKind {
			la = tmpl.ShutdownLastAction
		}

		// load params and snippets
		params := tmpl.KickstartParams{
			SystemID:   system.ID,
			ImageID:    *system.ImageID,
			LastAction: la,
			Snippets:   make(map[string][]string),
		}

		for _, kind := range model.AllSnippetKinds {
			snippets, err := sDao.FindByKind(ctx, system.ID, kind)
			if err != nil {
				return err
			}
			params.Snippets[kind.String()] = snippets
		}

		err = tmpl.RenderKickstartInstall(w, params)
	} else if system != nil && !system.Installable() {
		slog.WarnContext(ctx, "system found but not installable",
			"id", system.ID,
			"name", system.Name,
			"acquired_at", system.AcquiredAt.String(),
			"image_id", system.ImageID)
		err := tmpl.RenderKickstartDiscover(w)
		if err != nil {
			return err
		}
	} else {
		err := tmpl.RenderKickstartDiscover(w)
		if err != nil {
			return err
		}
	}

	return nil
}

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

	err = RenderKickstartForSystem(r.Context(), system, w)
	if err != nil {
		renderKsError(err, w, r)
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
