package ctl

import (
	"context"
	"fmt"
	"forester/internal/db"
	"forester/internal/model"
	"net"
	"time"
)

var _ SystemService = SystemServiceImpl{}

type SystemServiceImpl struct{}

func (i SystemServiceImpl) Register(ctx context.Context, system *NewSystem) error {
	dao := db.GetSystemDao(ctx)
	var hwAddrs []net.HardwareAddr
	for _, a := range system.HwAddrs {
		mac, err := net.ParseMAC(a)
		if err != nil {
			return fmt.Errorf("cannot parse hardware address '%s': %w", a, err)
		}
		hwAddrs = append(hwAddrs, mac)
	}
	var facts model.Facts
	for k, v := range system.Facts {
		facts.List = append(facts.List, model.Fact{Key: k, Value: v})
	}
	dbSystem := model.System{
		HwAddrs: hwAddrs,
		Facts:   facts,
	}

	err := dao.Register(ctx, &dbSystem)
	if err != nil {
		return fmt.Errorf("cannot create: %w", err)
	}

	return nil
}

func (i SystemServiceImpl) List(ctx context.Context, limit int64, offset int64) ([]*System, error) {
	dao := db.GetSystemDao(ctx)
	list, err := dao.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("cannot list: %w", err)
	}

	result := make([]*System, len(list))
	for i, item := range list {
		result[i] = &System{
			ID:         item.ID,
			Name:       item.Name,
			HwAddrs:    item.HwAddrStrings(),
			Facts:      item.FactsMap(),
			Acquired:   item.Acquired,
			AcquiredAt: item.AcquiredAt,
			ImageID:    item.ImageID,
			Comment:    item.Comment,
		}
	}

	return result, nil
}

func (i SystemServiceImpl) Acquire(ctx context.Context, system *SystemToAcquire) (time.Time, error) {
	dao := db.GetSystemDao(ctx)
	dbSystem := &model.System{
		ID:      system.ID,
		ImageID: &system.ImageID,
		Comment: system.Comment,
	}

	err := dao.Acquire(ctx, dbSystem)
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot acquire: %w", err)
	}

	return dbSystem.AcquiredAt, nil
}

func (i SystemServiceImpl) Release(ctx context.Context, systemId int64) error {
	dao := db.GetSystemDao(ctx)
	err := dao.Release(ctx, systemId)
	if err != nil {
		return fmt.Errorf("cannot release: %w", err)
	}

	return nil
}
