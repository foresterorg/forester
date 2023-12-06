package mux

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/db"
	"forester/internal/model"
	"forester/internal/tmpl"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/slog"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
)
import "github.com/go-chi/render"

func MountKickstart(r *chi.Mux) {
	r.Use(render.SetContentType(render.ContentTypePlainText))
	r.Use(DebugMiddleware)
	r.Get("/", HandleKickstart)
	r.Post("/register", HandleRegister)
}

var ErrMACHeaderInvalid = errors.New("invalid format of RHN MAC header")

func buildDiscoveryKickstartParams(ctx context.Context) (*tmpl.KickstartParams, error) {
	iDao := db.GetInstallationDao(ctx)
	i, s, err := iDao.FindInstallationForMAC(ctx, db.NullMAC)
	if err != nil {
		return nil, fmt.Errorf("no discovery system: %w", err)
	}

	result := tmpl.KickstartParams{
		ImageID:        0,
		SystemID:       s.ID,
		SystemName:     s.Name,
		SystemHostname: ToHostname(s.Name),
		InstallUUID:    i.UUID.String(),
		LastAction:     tmpl.ShutdownLastAction,
		Snippets:       make(map[string][]string),
	}

	nDao := db.GetSnippetDao(ctx)
	snippets, err := nDao.FindByKind(ctx, s.ID, model.PreSnippetKind)
	if err != nil {
		slog.ErrorContext(ctx, "error loading snippet", "id", s.ID, "kind", model.PreSnippetKind)
		return nil, err
	}
	result.Snippets[model.PreSnippetKind.String()] = snippets

	return &result, nil
}

func renderDiscover(ctx context.Context, w io.Writer) error {
	params, err := buildDiscoveryKickstartParams(ctx)
	if err != nil {
		return fmt.Errorf("error building discovery params: %w", err)
	}
	return tmpl.RenderKickstartDiscover(ctx, w, *params)
}

func RenderKickstartForSystem(ctx context.Context, system *model.System, w io.Writer) error {
	if system == nil {
		slog.DebugContext(ctx, "no system found, missing Anaconda MAC header")
		return renderDiscover(ctx, w)
	}

	inDao := db.GetInstallationDao(ctx)
	insts, err := inDao.FindValidByState(ctx, system.ID, model.InstallingInstallState)
	var inst *model.Installation
	if err != nil {
		slog.ErrorContext(ctx, "error during finding installations for a system", "id", system.ID, "err", err)
		return renderDiscover(ctx, w)
	}

	if len(insts) == 0 {
		slog.WarnContext(ctx, "system found but not installable",
			"id", system.ID,
			"name", system.Name,
			"acquired_at", system.AcquiredAt.String())
		return renderDiscover(ctx, w)
	}
	inst = insts[0]

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

	// load associated image
	var liveimgSha256 string
	iDao := db.GetImageDao(ctx)
	img, err := iDao.FindByID(ctx, inst.ImageID)
	if err != nil {
		slog.ErrorContext(ctx, "error loading image for system", "id", system.ID, "image_id", inst.ImageID)
		return err
	}
	liveimgSha256 = img.LiveimgSha256

	// load params and snippets
	params := tmpl.KickstartParams{
		SystemID:       system.ID,
		ImageID:        inst.ImageID,
		SystemName:     system.Name,
		SystemHostname: ToHostname(system.Name),
		InstallUUID:    inst.UUID.String(),
		LastAction:     la,
		Snippets:       make(map[string][]string),
		CustomSnippet:  system.CustomSnippet,
		LiveimgSha256:  liveimgSha256,
	}

	for _, kind := range model.AllSnippetKinds {
		snippets, err := sDao.FindByKind(ctx, system.ID, kind)
		if err != nil {
			slog.ErrorContext(ctx, "error loading snippet", "id", system.ID, "kind", kind)
			return err
		}
		params.Snippets[kind.String()] = snippets
	}

	err = tmpl.RenderKickstartInstall(ctx, w, params)
	if err != nil {
		slog.ErrorContext(ctx, "error rendering ks snippet", "id", system.ID)
		return err
	}

	return nil
}

var headerRegexp = regexp.MustCompile("(?i)^X-RHN-Provisioning-MAC-")

func HandleKickstart(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	var system *model.System
	var err error

	sDao := db.GetSystemDao(r.Context())
	for k, v := range r.Header {
		if !headerRegexp.MatchString(k) {
			slog.DebugContext(r.Context(), "skipping header", "name", k, "value", v)
			continue
		}
		for _, macString := range v {
			slog.DebugContext(r.Context(), "processing header", "name", k, "value", v)
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
	err := tmpl.RenderKickstartError(r.Context(), w, tmpl.KickstartErrorParams{Message: ksErr.Error()})
	if err != nil {
		slog.ErrorContext(r.Context(), "cannot render template", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
