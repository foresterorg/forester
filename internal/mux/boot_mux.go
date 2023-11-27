package mux

import (
	"errors"
	"forester/internal/config"
	"forester/internal/db"
	"forester/internal/model"
	"forester/internal/tmpl"
	"net"
	"net/http"
	"strings"

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
	var headers []slog.Attr
	for k, v := range r.Header {
		pair := slog.String(k, strings.Join(v, " "))
		headers = append(headers, pair)
	}
	if len(headers) > 0 {
		slog.DebugContext(r.Context(), "HTTP headers", "headers", slog.GroupValue(headers...))
	}
	fs := http.StripPrefix("/boot", http.FileServer(http.Dir(config.BootPath())))
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

func HandleMacConfig(w http.ResponseWriter, r *http.Request) {
	mac, err := net.ParseMAC(chi.URLParam(r, "MAC"))
	if err != nil {
		renderGrubError(err, w, r)
		return
	}

	var imageId, systemId int64
	inst := &model.Installation{}

	sDao := db.GetSystemDao(r.Context())
	system, err := sDao.FindByMac(r.Context(), mac)

	if errors.Is(err, pgx.ErrNoRows) {
		slog.InfoContext(r.Context(), "unknown system - booting discovery", "mac", mac.String())
		imageId = config.Images.BootId
	} else if err != nil {
		slog.ErrorContext(r.Context(), "error while finding system", "mac", mac.String(), "err", err)
		renderGrubError(err, w, r)
		return
	} else {
		systemId = system.ID
		iDao := db.GetInstallationDao(r.Context())
		insts, err := iDao.FindValidByState(r.Context(), systemId, model.FinishedInstallState)
		if err != nil {
			slog.ErrorContext(r.Context(), "cannot find installations for system", "mac", mac.String(), "err", err)
			renderGrubError(err, w, r)
			return
		}

		if len(insts) > 0 {
			slog.InfoContext(r.Context(), "known system - booting installer", "mac", mac.String(), "pending_installs", len(insts), "install_id", inst.ID, "install_uuid", inst.UUID)
			inst = insts[0]
			imageId = inst.ImageID
		} else {
			slog.InfoContext(r.Context(), "known system but not installable - booting discovery", "mac", mac.String())
			imageId = config.Images.BootId
		}
	}

	w.WriteHeader(http.StatusOK)
	err = tmpl.RenderGrubKernel(r.Context(), w, tmpl.GrubKernelParams{ImageID: imageId, SystemID: systemId, InstallUUID: inst.UUID.String()})
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
