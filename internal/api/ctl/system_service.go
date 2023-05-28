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
