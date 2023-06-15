package ctl

import "github.com/go-chi/chi/v5"

var Service = struct {
	Image     ImageService
	Appliance ApplianceService
	System    SystemService
}{
	ImageServiceImpl{},
	ApplianceServiceImpl{},
	SystemServiceImpl{},
}

func MountServices(r chi.Router) {
	imageSrvHandler := NewImageServiceServer(Service.Image)
	r.Handle("/rpc/ImageService/*", imageSrvHandler)
	applianceSrvHandler := NewApplianceServiceServer(Service.Appliance)
	r.Handle("/rpc/ApplianceService/*", applianceSrvHandler)
	systemSrvHandler := NewSystemServiceServer(Service.System)
	r.Handle("/rpc/SystemService/*", systemSrvHandler)
}
