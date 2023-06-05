package ctl

import (
	"context"
	"fmt"
	"forester/internal/db"
	"forester/internal/model"
	"net"
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

func (i SystemServiceImpl) Find(ctx context.Context, pattern string) (*System, error) {
	dao := db.GetSystemDao(ctx)
	result, err := dao.Find(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("cannot find: %w", err)
	}

	hwa := make([]string, len(result.HwAddrs))
	for i, _ := range result.HwAddrs {
		hwa[i] = result.HwAddrs[i].String()
	}

	return &System{
		ID:         result.ID,
		Name:       result.Name,
		HwAddrs:    hwa,
		Facts:      result.Facts.FactsMap(),
		Acquired:   result.Acquired,
		AcquiredAt: result.AcquiredAt,
		ImageID:    result.ImageID,
		Comment:    result.Comment,
	}, nil
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
			Facts:      item.Facts.FactsMap(),
			Acquired:   item.Acquired,
			AcquiredAt: item.AcquiredAt,
			ImageID:    item.ImageID,
			Comment:    item.Comment,
		}
	}

	return result, nil
}

func (i SystemServiceImpl) Acquire(ctx context.Context, systemPattern, imagePattern, comment string) error {
	daoSystem := db.GetSystemDao(ctx)
	daoImage := db.GetImageDao(ctx)

	image, err := daoImage.Find(ctx, imagePattern)
	if err != nil {
		return fmt.Errorf("cannot find: %w", err)
	}
	system, err := daoSystem.Find(ctx, systemPattern)
	if err != nil {
		return fmt.Errorf("cannot find: %w", err)
	}

	err = daoSystem.Acquire(ctx, system.ID, image.ID, comment)
	if err != nil {
		return fmt.Errorf("cannot acquire: %w", err)
	}

	return nil
}

func (i SystemServiceImpl) Release(ctx context.Context, systemPattern string) error {
	dao := db.GetSystemDao(ctx)

	system, err := dao.Find(ctx, systemPattern)
	if err != nil {
		return fmt.Errorf("cannot find: %w", err)
	}

	err = dao.Release(ctx, system.ID)
	if err != nil {
		return fmt.Errorf("cannot release: %w", err)
	}

	return nil
}
