package metal

import (
	"context"
	"errors"

	"forester/internal/model"
)

type Metal interface {
	Enlist(ctx context.Context, app *model.Appliance, pattern string) ([]*EnlistResult, error)
	BootNetwork(ctx context.Context, system *model.SystemAppliance) error
	BootLocal(ctx context.Context, system *model.SystemAppliance) error
}

type EnlistResult struct {
	HwAddrs []string          `json:"HwAddrs"`
	Facts   map[string]string `json:"Facts"`
	UID     string            `json:"UID"`
}

var noopMetal Metal = NoopMetal{}

var libvirtMetal Metal = LibvirtMetal{}

var redfishMetal Metal = RedfishMetal{}

var redfishManualMetal Metal = RedfishMetal{Manual: true}

func ForKind(kind model.ApplianceKind) Metal {
	switch kind {
	case model.LibvirtApplianceKind:
		return libvirtMetal
	case model.RedfishApplianceKind:
		return redfishMetal
	case model.RedfishManualApplianceKind:
		return redfishManualMetal
	}
	return noopMetal
}

func Enlist(ctx context.Context, app *model.Appliance, pattern string) ([]*EnlistResult, error) {
	metal := ForKind(app.Kind)
	return metal.Enlist(ctx, app, pattern)
}

var ErrSystemWithNoAppliance = errors.New("system has no appliance associated")

var ErrSystemWithNoUID = errors.New("system has no UID set")

func BootNetwork(ctx context.Context, system *model.SystemAppliance) error {
	if system.ApplianceID == nil {
		return ErrSystemWithNoAppliance
	}

	if system.UID == nil {
		return ErrSystemWithNoUID
	}

	metal := ForKind(system.Appliance.Kind)
	return metal.BootNetwork(ctx, system)
}

func BootLocal(ctx context.Context, system *model.SystemAppliance) error {
	if system.ApplianceID == nil {
		return ErrSystemWithNoAppliance
	}

	if system.UID == nil {
		return ErrSystemWithNoUID
	}

	metal := ForKind(system.Appliance.Kind)
	return metal.BootLocal(ctx, system)
}
