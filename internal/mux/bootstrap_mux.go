package mux

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"

	"forester/internal/tmpl"
)

func MountBootstrap(r *chi.Mux) {
	r.Head("/ipxe/chain.ipxe", serveIpxeBootstrap)
	r.Get("/ipxe/chain.ipxe", serveIpxeBootstrap)
	r.Head("/ipxe/*", serveIpxeBootstrapFile)
	r.Get("/ipxe/*", serveIpxeBootstrapFile)
}

func serveIpxeBootstrapFile(w http.ResponseWriter, r *http.Request) {
	fs := http.StripPrefix("/bootstrap/ipxe", http.FileServer(http.Dir("/usr/share/ipxe")))
	fs.ServeHTTP(w, r)
}

func serveIpxeBootstrap(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	err := tmpl.RenderIpxeBootstrap(r.Context(), w)
	if err != nil {
		renderBootError(err, w, r, tmpl.IpxeBootErrorType)
		return
	}
}
