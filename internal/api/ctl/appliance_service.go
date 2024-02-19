package ctl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"forester/internal/db"
	"forester/internal/metal"
	"forester/internal/model"
)

var _ ApplianceService = ApplianceServiceImpl{}

type ApplianceServiceImpl struct{}

var ErrUnknownApplianceKind = errors.New("unknown appliance kind")

func (i ApplianceServiceImpl) Create(ctx context.Context, name string, kind int16, uri string) error {
	dao := db.GetApplianceDao(ctx)
	record := model.Appliance{
		Kind: model.ParseKind(kind),
		Name: name,
		URI:  uri,
	}
	if record.Kind < 0 {
		return fmt.Errorf("%w: %d", ErrUnknownApplianceKind, kind)
	}

	err := dao.Create(ctx, &record)
	if err != nil {
		return fmt.Errorf("cannot create: %w", err)
	}

	return nil
}

func (i ApplianceServiceImpl) Find(ctx context.Context, name string) (*Appliance, error) {
	dao := db.GetApplianceDao(ctx)
	result, err := dao.Find(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("cannot find: %w", err)
	}

	return &Appliance{
		ID:   result.ID,
		Name: result.Name,
		Kind: int16(result.Kind),
		URI:  result.URI,
	}, nil
}

func (i ApplianceServiceImpl) List(ctx context.Context, limit int64, offset int64) ([]*Appliance, error) {
	dao := db.GetApplianceDao(ctx)
	ensureLimitNonzero(&limit)
	list, err := dao.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("cannot list: %w", err)
	}

	result := make([]*Appliance, len(list))
	for i, item := range list {
		result[i] = &Appliance{
			ID:   item.ID,
			Name: item.Name,
			Kind: int16(item.Kind),
			URI:  item.URI,
		}
	}

	return result, nil
}

func (i ApplianceServiceImpl) Enlist(ctx context.Context, name string, namePattern string) error {
	dao := db.GetApplianceDao(ctx)
	app, err := dao.Find(ctx, name)
	if err != nil {
		return fmt.Errorf("unknown appliance '%s': %w", name, err)
	}
	systems, err := metal.Enlist(ctx, app, namePattern)
	if err != nil {
		slog.ErrorContext(ctx, "error when enlisting", "err", err.Error())
		return fmt.Errorf("cannot enlist: %w", err)
	}

	for _, system := range systems {
		ns := &NewSystem{
			HwAddrs:       system.HwAddrs,
			Facts:         system.Facts,
			ApplianceName: &app.Name,
			UID:           &system.UID,
		}

		slog.InfoContext(ctx, "registering system",
			"mac", strings.Join(ns.HwAddrs, ","),
			"uuid", system.UID,
			"appliance", app.Name,
		)
		err = Service.System.Register(ctx, ns)
		if err != nil {
			return fmt.Errorf("cannot register system: %w", err)
		}
	}

	return nil
}

func (i ApplianceServiceImpl) Delete(ctx context.Context, name string) error {
	panic("implement me")
}
