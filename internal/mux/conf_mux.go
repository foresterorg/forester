package mux

import (
	"encoding/hex"
	"log/slog"
	"net/http"

	chi "github.com/go-chi/chi/v5"

	"forester/internal/db"
	"forester/internal/tmpl"
)

func MountConf(r *chi.Mux) {
	r.Get("/iscdhcpd/grub", serveConf("iscdhcpd", "grub"))
	r.Get("/iscdhcpd/ipxe", serveConf("iscdhcpd", "ipxe"))
	r.Get("/dnsmasq/grub", serveConf("dnsmasq", "grub"))
	r.Get("/dnsmasq/ipxe", serveConf("dnsmasq", "ipxe"))
	r.Get("/libvirt/grub", serveConf("libvirt", "grub"))
	r.Get("/libvirt/ipxe", serveConf("libvirt", "ipxe"))
}

func allZero(s []byte) bool {
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return true
}

func serveConf(name, format string) func(w http.ResponseWriter, r *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sDao := db.GetSystemDao(ctx)
		systems, err := sDao.List(ctx, 1000000, 0)
		if err != nil {
			slog.WarnContext(ctx, "error during dnsmasq config generation", "err", err)
			http.Error(w, "# system list error: ", http.StatusInternalServerError)
			return
		}

		entries := make([]tmpl.DhcpEntry, 0, len(systems)*4)
		for _, s := range systems {
			for _, mac := range s.HwAddrs.Unique() {
				if allZero(mac) {
					continue
				}
				e := tmpl.DhcpEntry{
					Tag: "t" + hex.EncodeToString(mac),
					MAC: mac.String(),
				}
				entries = append(entries, e)
			}
		}

		err = tmpl.RenderDhcpConf(r.Context(), w, name, format, tmpl.DhcpParams{Entries: entries})
		if err != nil {
			slog.ErrorContext(r.Context(), "cannot render dhcp config template", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	return f
}
