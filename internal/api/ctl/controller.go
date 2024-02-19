package ctl

import chi "github.com/go-chi/chi/v5"

var Service = struct {
	Image     ImageService
	Appliance ApplianceService
	System    SystemService
	Snippet   SnippetService
}{
	ImageServiceImpl{},
	ApplianceServiceImpl{},
	SystemServiceImpl{},
	SnippetServiceImpl{},
}

func MountServices(r chi.Router) {
	imageSrvHandler := NewImageServiceServer(Service.Image)
	r.Handle("/rpc/ImageService/*", imageSrvHandler)
	applianceSrvHandler := NewApplianceServiceServer(Service.Appliance)
	r.Handle("/rpc/ApplianceService/*", applianceSrvHandler)
	systemSrvHandler := NewSystemServiceServer(Service.System)
	r.Handle("/rpc/SystemService/*", systemSrvHandler)
	snippetSrvHandler := NewSnippetServiceServer(Service.Snippet)
	r.Handle("/rpc/SnippetService/*", snippetSrvHandler)
}
