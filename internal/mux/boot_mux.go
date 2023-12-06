package mux

import (
	"context"
	"fmt"
	"forester/internal/config"
	"forester/internal/db"
	"forester/internal/model"
	"forester/internal/tmpl"
	"io"
	"mime"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
)

func MountBoot(r *chi.Mux) {
	paths := []string{
		"/shim.efi",
		"/grubx64.efi",
		"//grubx64.efi", // some grub versions request double slash
		"/.discinfo",
		"/LICENSE",
		"/liveimg.tar.gz",
		"/images/*",
		"/x86_64-efi/*",
		"/boot.iso",
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
		// sourced grub config
		r.Head("/grub.cfg/{MAC}", HandleMacConfig)
		r.Get("/grub.cfg/{MAC}", HandleMacConfig)
		// redhat patched grub config
		r.Head("/grub.cfg-{MAC}", HandleMacConfig)
		r.Get("/grub.cfg-{MAC}", HandleMacConfig)
		// managed grub config
		r.Head("/{MAC}/grub.cfg", HandleMacConfig)
		r.Get("/{MAC}/grub.cfg", HandleMacConfig)
	})

	mime.AddExtensionType(".iso", "application/vnd.efi.iso")
	mime.AddExtensionType(".img", "application/vnd.efi.img")
	mime.AddExtensionType(".efi", "application/efi")
}

func serveBootPath(w http.ResponseWriter, r *http.Request) {
	var err error
	var s *model.System
	var i *model.Installation

	origMAC := chi.URLParam(r, "MAC")
	mac, _ := net.ParseMAC(origMAC)

	iDao := db.GetInstallationDao(r.Context())
	i, s, err = iDao.FindInstallationForMAC(r.Context(), mac)
	if err != nil {
		slog.WarnContext(r.Context(), "not found", "mac", mac.String(), "err", err)
		http.NotFound(w, r)
		return
	}

	root := config.BootPath(i.ImageID)
	prefix := "/boot"
	if origMAC != "" {
		prefix = fmt.Sprintf("/boot/%s", origMAC)
	}
	slog.InfoContext(r.Context(), "serving root", "directory", root, "system_id", s.ID, "install_uuid", i.UUID, "path", r.URL.Path, "prefix", prefix)
	fs := http.StripPrefix(prefix, http.FileServer(http.Dir(root)))
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
	origMAC := chi.URLParam(r, "MAC")
	mac, _ := net.ParseMAC(origMAC)

	err := WriteMacConfig(r.Context(), w, mac)
	if err != nil {
		renderGrubError(err, w, r)
		return
	}
}

func WriteMacConfig(ctx context.Context, w io.Writer, mac net.HardwareAddr) error {
	var err error
	var s *model.System
	var i *model.Installation

	iDao := db.GetInstallationDao(ctx)
	i, s, err = iDao.FindInstallationForMAC(ctx, mac)
	if err != nil {
		return err
	}

	params := tmpl.GrubKernelParams{
		SystemID:    s.ID,
		ImageID:     i.ImageID,
		InstallUUID: i.UUID.String(),
	}

	err = tmpl.RenderGrubKernel(ctx, w, params)
	if err != nil {
		return err
	}

	return nil
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
