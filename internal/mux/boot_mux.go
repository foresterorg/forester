package mux

import (
	"context"
	"errors"
	"forester/internal/config"
	"forester/internal/db"
	"forester/internal/model"
	"forester/internal/tmpl"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/slog"
)

func MountBoot(r *chi.Mux) {
	paths := []string{
		"/shim.efi",
		"/grubx64.efi",
		"//grubx64.efi", // some grub versions request double slash
		"/.discinfo",
		"/liveimg.tar.gz",
		"/images/*",
	}

	for _, path := range paths {
		// anonymous bootstrap paths
		r.Head(path, serveBootPath)
		r.Get(path, serveBootPath)

		// managed bootstrap paths
		r.Head("/{MAC}"+path, serveBootPath)
		r.Get("/{MAC}"+path, serveBootPath)
	}

	r.Group(func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypePlainText))
		r.Use(DebugMiddleware)

		// anonymous grub config
		r.Head("/grub.cfg", HandleBootstrapConfig)
		r.Get("/grub.cfg", HandleBootstrapConfig)
		// managed grub config
		r.Head("/grub.cfg/{MAC}", HandleMacConfig)
		r.Get("/grub.cfg/{MAC}", HandleMacConfig)
		// redhat patched grub config
		r.Head("/grub.cfg-{MAC}", HandleMacConfig)
		r.Get("/grub.cfg-{MAC}", HandleMacConfig)
	})
}

func serveBootPath(w http.ResponseWriter, r *http.Request) {
	var s *model.System
	var i *model.Installation

	mac, err := net.ParseMAC(chi.URLParam(r, "MAC"))
	if err != nil {
		s, i = findDiscoveryInstall(r.Context())
	} else {
		sDao := db.GetSystemDao(r.Context())
		s, err = sDao.FindByMac(r.Context(), mac)
		if err != nil {
			slog.WarnContext(r.Context(), "cannot find host by mac", "mac", mac, "parsed_mac", mac.String())
			s, i = findDiscoveryInstall(r.Context())
		} else {
			iDao := db.GetInstallationDao(r.Context())
			is, err := iDao.FindValidByState(r.Context(), s.ID, model.AnyInstallState)
			if err != nil || len(is) < 1 {
				slog.WarnContext(r.Context(), "system has no active installation", "mac", mac.String())
				s, i = findDiscoveryInstall(r.Context())
			} else {
				if len(is) > 1 {
					slog.WarnContext(r.Context(), "more than one installations for host", "mac", mac.String())
				} else {
					slog.WarnContext(r.Context(), "found installation for system", "mac", mac.String(), "system_id", s.ID)
				}
				i = is[0]
			}
		}
	}

	root := config.BootPath(i.ImageID)
	if s.ID == 0 {
		slog.WarnContext(r.Context(), "cannot find system or discovery system, bootstrap will fail")
	}
	slog.InfoContext(r.Context(), "serving root", "directory", root, "system_id", s.ID, "install_uuid", i.UUID, "path", r.URL.Path)
	fs := http.StripPrefix("/boot", http.FileServer(http.Dir(root)))
	fs.ServeHTTP(w, r)
}

func HandleBootstrapConfig(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	err := tmpl.RenderGrubBootstrap(r.Context(), w)
	if err != nil {
		renderGrubError(err, w, r)
		return
	}
}

func buildDiscoveryGrubParams(ctx context.Context) tmpl.GrubKernelParams {
	s, i := findDiscoveryInstall(ctx)

	result := tmpl.GrubKernelParams{
		ImageID:     i.ImageID,
		SystemID:    s.ID,
		InstallUUID: i.UUID.String(),
	}
	return result
}

func HandleMacConfig(w http.ResponseWriter, r *http.Request) {
	mac, err := net.ParseMAC(chi.URLParam(r, "MAC"))
	if err != nil {
		renderGrubError(err, w, r)
		return
	}

	params := tmpl.GrubKernelParams{}

	sDao := db.GetSystemDao(r.Context())
	system, err := sDao.FindByMac(r.Context(), mac)

	if errors.Is(err, pgx.ErrNoRows) {
		slog.InfoContext(r.Context(), "unknown system - booting discovery", "mac", mac.String())
		params = buildDiscoveryGrubParams(r.Context())
	} else if err != nil {
		slog.ErrorContext(r.Context(), "error while finding system", "mac", mac.String(), "err", err)
		renderGrubError(err, w, r)
		return
	} else {
		params.SystemID = system.ID
		iDao := db.GetInstallationDao(r.Context())
		insts, err := iDao.FindValidByState(r.Context(), params.SystemID, model.FinishedInstallState)
		if err != nil {
			slog.ErrorContext(r.Context(), "cannot find installations for system", "mac", mac.String(), "err", err)
			renderGrubError(err, w, r)
			return
		}

		if len(insts) > 0 {
			slog.InfoContext(r.Context(), "known system - booting installer", "mac", mac.String(), "pending_installs", len(insts), "image_id", params.ImageID, "install_uuid", params.InstallUUID)
			params.ImageID = insts[0].ImageID
			params.InstallUUID = insts[0].UUID.String()
		} else {
			slog.InfoContext(r.Context(), "known system but not installable - booting discovery", "mac", mac.String())
			params = buildDiscoveryGrubParams(r.Context())
		}
	}

	w.WriteHeader(http.StatusOK)
	err = tmpl.RenderGrubKernel(r.Context(), w, params)
	if err != nil {
		renderGrubError(err, w, r)
		return
	}
}

func renderGrubError(gerr error, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	slog.ErrorContext(r.Context(), "rendering error as grub message", "err", gerr)
	err := tmpl.RenderGrubError(r.Context(), w, tmpl.GrubErrorParams{Error: gerr})
	if err != nil {
		slog.ErrorContext(r.Context(), "cannot render template", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
