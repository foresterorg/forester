package ctl

import "github.com/go-chi/chi/v5"

func MountServices(r chi.Router) {
	imageSrvHandler := NewImageServiceServer(ImageServiceImpl{})
	r.Handle("/rpc/ImageService/*", imageSrvHandler)
	applianceSrvHandler := NewApplianceServiceServer(ApplianceServiceImpl{})
	r.Handle("/rpc/ApplianceService/*", applianceSrvHandler)
	systemSrvHandler := NewSystemServiceServer(SystemServiceImpl{})
	r.Handle("/rpc/SystemService/*", systemSrvHandler)
}
