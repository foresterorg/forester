package mux

import (
	"encoding/hex"
	"fmt"
	"forester/internal/config"
	"forester/internal/db"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
)

func MountConf(r *chi.Mux) {
	r.Get("/dnsmasq", serveDnsmasq)
}

func printDnsmasqHeader(w http.ResponseWriter) {
	fmt.Fprintln(w, "#\n#")
	fmt.Fprintln(w, "# Place this config in /etc/dnsmasq.d/forester.conf and restart dnsmasq.")
	fmt.Fprintln(w, "#\n#")
	fmt.Fprintln(w, "dhcp-vendorclass=set:bios,PXEClient:Arch:00000")
	fmt.Fprintln(w, "dhcp-vendorclass=set:efi,PXEClient:Arch:00007")
	fmt.Fprintln(w, "dhcp-vendorclass=set:efix64,PXEClient:Arch:00009")
	fmt.Fprintln(w, "dhcp-vendorclass=set:efihttp,HTTPClient:Arch:00016")
	fmt.Fprintln(w, "dhcp-option-force=tag:efihttp,60,HTTPClient")
}

func serveDnsmasq(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sDao := db.GetSystemDao(ctx)
	systems, err := sDao.List(ctx, 1000000, 0)
	if err != nil {
		slog.WarnContext(ctx, "error during dnsmasq config generation", "err", err)
		http.Error(w, "# system list error: ", http.StatusInternalServerError)
		return
	}
	printDnsmasqHeader(w)
	for _, s := range systems {
		fmt.Fprintln(w, "\n# "+s.Name)
		for _, mac := range s.HwAddrs.Unique() {
			tag := "t" + hex.EncodeToString(mac)
			if tag == "t000000000000" {
				continue
			}
			fmt.Fprintf(w, "dhcp-host=%s,set:%s\n", mac, tag)
			fmt.Fprintf(w, "dhcp-boot=tag:bios,tag:%s,boot/bios/%s/grubx64.0,,%s\n", tag, mac, config.BaseHost())
			fmt.Fprintf(w, "dhcp-boot=tag:efi,tag:%s,boot/efi/%s/shim.efi,,%s\n", tag, mac, config.BaseHost())
			fmt.Fprintf(w, "dhcp-boot=tag:efi64,tag:%s,boot/efi64/%s/shim.efi,,%s\n", tag, mac, config.BaseHost())
			fmt.Fprintf(w, "dhcp-boot=tag:efihttp,tag:%s,%s/boot/efi64/%s/shim.efi\n", tag, config.BaseURL(), mac)
		}
	}
}
