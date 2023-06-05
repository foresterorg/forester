package mux

import (
	"errors"
	"forester/internal/config"
	"forester/internal/db"
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
		r.Head(path, serveBootPath)
		r.Get(path, serveBootPath)
	}

	r.Group(func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypePlainText))
		r.Use(DebugMiddleware)

		r.Head("/grub.cfg", HandleBootstrapConfig)
		r.Get("/grub.cfg", HandleBootstrapConfig)
		r.Head("/mac/{MAC}", HandleMacConfig)
		r.Get("/mac/{MAC}", HandleMacConfig)
	})
}

func serveBootPath(w http.ResponseWriter, r *http.Request) {
	fs := http.StripPrefix("/boot", http.FileServer(http.Dir(config.BootPath())))
	fs.ServeHTTP(w, r)
}

func HandleBootstrapConfig(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	err := tmpl.RenderGrubBootstrap(w)
	if err != nil {
		renderGrubError(err, w, r)
		return
	}
}

var ErrSystemNotInstallable = errors.New("system is not installable, acquire it again or change APP_INSTALL_DURATION")

func HandleMacConfig(w http.ResponseWriter, r *http.Request) {
	mac, err := net.ParseMAC(chi.URLParam(r, "MAC"))
	if err != nil {
		renderGrubError(err, w, r)
		return
	}

	var imageId int64

	sDao := db.GetSystemDao(r.Context())
	system, err := sDao.FindByMac(r.Context(), mac)
	if errors.Is(err, pgx.ErrNoRows) {
		// unknown system: boot discovery anaconda
		imageId = config.Images.BootId
	} else if err != nil {
		renderGrubError(err, w, r)
		return
	} else {
		// known system
		if !system.Installable() || system.ImageID == nil {
			renderGrubError(ErrSystemNotInstallable, w, r)
			return
		}
		imageId = *system.ImageID
	}

	w.WriteHeader(http.StatusOK)
	err = tmpl.RenderGrubKernel(w, tmpl.GrubKernelParams{ImageID: imageId})
	if err != nil {
		renderGrubError(err, w, r)
		return
	}
}

func renderGrubError(gerr error, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	slog.ErrorCtx(r.Context(), "grub template error", "err", gerr)
	err := tmpl.RenderGrubError(w, tmpl.GrubErrorParams{Error: gerr})
	if err != nil {
		slog.ErrorCtx(r.Context(), "cannot render template", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
