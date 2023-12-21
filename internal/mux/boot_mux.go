package mux

import (
	"context"
	"forester/internal/config"
	"forester/internal/db"
	"forester/internal/model"
	"forester/internal/tmpl"
	"io"
	"mime"
	"net"
	"net/http"
	"slices"
	"strings"

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
		"/boot.iso",
	}

	for _, path := range paths {
		r.Head("/{PLATFORM}/{MAC}"+path, serveBootPath)
		r.Get("/{PLATFORM}/{MAC}"+path, serveBootPath)
	}

	r.Head("/ipxe/*", serveIpxeEFI)
	r.Get("/ipxe/*", serveIpxeEFI)
	r.Head("/ipxes/{MAC}/*", serveIpxeScript)
	r.Get("/ipxes/{MAC}/*", serveIpxeScript)

	r.Group(func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypePlainText))
		r.Use(DebugMiddleware)

		r.Head("/{PLATFORM}/{MAC}/grub.cfg", HandleMacConfig)
		r.Get("/{PLATFORM}/{MAC}/grub.cfg", HandleMacConfig)
	})

	mime.AddExtensionType(".iso", "application/vnd.efi.iso")
	mime.AddExtensionType(".img", "application/vnd.efi.img")
	mime.AddExtensionType(".efi", "application/efi")
}

func renderBootError(gerr error, w http.ResponseWriter, r *http.Request, t tmpl.BootErrorType) {
	slog.ErrorContext(r.Context(), "boot error", "err", gerr)
	w.WriteHeader(http.StatusOK)
	err := tmpl.RenderBootError(r.Context(), w, tmpl.BootErrorParams{Type: t, Error: gerr})
	if err != nil {
		slog.ErrorContext(r.Context(), "cannot render boot error template", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func serveBootPath(w http.ResponseWriter, r *http.Request) {
	var err error
	var s *model.System
	var i *model.Installation

	platform := chi.URLParam(r, "PLATFORM")
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

	prefix := "/" + strings.Join(slices.DeleteFunc([]string{"boot", platform, origMAC}, func(e string) bool {
		return e == ""
	}), "/")
	slog.InfoContext(r.Context(), "serving root",
		"directory", root,
		"system_id", s.ID,
		"install_uuid", i.UUID,
		"path", r.URL.Path,
		"prefix", prefix,
	)
	fs := http.StripPrefix(prefix, http.FileServer(http.Dir(root)))
	fs.ServeHTTP(w, r)
}

func HandleBootstrapConfig(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	err := tmpl.RenderGrubBootstrap(r.Context(), w)
	if err != nil {
		renderBootError(err, w, r, tmpl.GrubBootErrorType)
		return
	}
}

func HandleMacConfig(w http.ResponseWriter, r *http.Request) {
	platform := strings.ToLower(chi.URLParam(r, "PLATFORM"))
	origMAC := chi.URLParam(r, "MAC")
	mac, _ := net.ParseMAC(origMAC)

	var linux tmpl.GrubLinuxCmd
	var initrd tmpl.GrubInitrdCmd
	if platform == "bios" {
		linux = tmpl.GrubLinuxCmdBIOS
		initrd = tmpl.GrubInitrdCmdBIOS
	} else {
		linux = tmpl.GrubLinuxCmdEFIX64
		initrd = tmpl.GrubInitrdCmdEFIX64
	}

	err := WriteGrubConfig(r.Context(), w, mac, linux, initrd)
	if err != nil {
		renderBootError(err, w, r, tmpl.GrubBootErrorType)
		return
	}
}

func WriteGrubConfig(ctx context.Context, w io.Writer, mac net.HardwareAddr, linux tmpl.GrubLinuxCmd, initrd tmpl.GrubInitrdCmd) error {
	var err error
	var s *model.System
	var i *model.Installation

	iDao := db.GetInstallationDao(ctx)
	i, s, err = iDao.FindInstallationForMAC(ctx, mac)
	if err != nil {
		return err
	}

	params := tmpl.BootKernelParams{
		SystemID:    s.ID,
		ImageID:     i.ImageID,
		InstallUUID: i.UUID.String(),
		LinuxCmd:    linux,
		InitrdCmd:   initrd,
	}

	err = tmpl.RenderGrubKernel(ctx, w, params)
	if err != nil {
		return err
	}

	return nil
}

func serveIpxeEFI(w http.ResponseWriter, r *http.Request) {
	fs := http.StripPrefix("/boot/ipxe", http.FileServer(http.Dir("/usr/share/ipxe")))
	fs.ServeHTTP(w, r)
}

func serveIpxeScript(w http.ResponseWriter, r *http.Request) {
	origMAC := chi.URLParam(r, "MAC")
	mac, _ := net.ParseMAC(origMAC)

	err := WriteIpxeConfig(r.Context(), w, mac)
	if err != nil {
		renderBootError(err, w, r, tmpl.IpxeBootErrorType)
		return
	}
}

func WriteIpxeConfig(ctx context.Context, w io.Writer, mac net.HardwareAddr) error {
	var err error
	var s *model.System
	var i *model.Installation

	iDao := db.GetInstallationDao(ctx)
	i, s, err = iDao.FindInstallationForMAC(ctx, mac)
	if err != nil {
		return err
	}

	params := tmpl.BootKernelParams{
		SystemID:    s.ID,
		ImageID:     i.ImageID,
		InstallUUID: i.UUID.String(),
	}

	err = tmpl.RenderIpxeKernel(ctx, w, params)
	if err != nil {
		return err
	}

	return nil
}
