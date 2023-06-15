package ctl

import (
	"context"
	"encoding/xml"
	"fmt"
	"forester/internal/db"
	"forester/internal/model"
	"net"

	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"github.com/google/uuid"
	"libvirt.org/go/libvirtxml"
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

	if system.ApplianceName != "" {
		ad := db.GetApplianceDao(ctx)
		app, err := ad.Find(ctx, system.ApplianceName)
		if err != nil {
			return fmt.Errorf("cannot find appliance named '%s': %w", system.ApplianceName, err)
		}
		dbSystem.ApplianceID = &app.ID
	}

	if system.UID != "" {
		dbSystem.UID = &system.UID
	}

	err := dao.Register(ctx, &dbSystem)
	if err != nil {
		return fmt.Errorf("cannot create: %w", err)
	}

	return nil
}

func (i SystemServiceImpl) Find(ctx context.Context, pattern string) (*System, error) {
	dao := db.GetSystemDao(ctx)
	result, err := dao.FindRelated(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("cannot find: %w", err)
	}

	hwa := make([]string, len(result.System.HwAddrs))
	for i, _ := range result.System.HwAddrs {
		hwa[i] = result.System.HwAddrs[i].String()
	}

	payload := &System{
		ID:         result.System.ID,
		Name:       result.System.Name,
		HwAddrs:    hwa,
		Facts:      result.System.Facts.FactsMap(),
		Acquired:   result.System.Acquired,
		AcquiredAt: result.System.AcquiredAt,
		ImageID:    result.System.ImageID,
		Comment:    result.System.Comment,
		UID:        result.System.UID,
	}

	payload.Appliance = &Appliance{
		ID:   result.Appliance.ID,
		Name: result.Appliance.Name,
		Kind: int16(result.Appliance.Kind),
		URI:  result.Appliance.URI,
	}

	return payload, nil
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

func (i SystemServiceImpl) Reset(ctx context.Context, systemPattern string) error {
	//daoApp := db.GetApplianceDao(ctx)
	//daoApp.Find(ctx, sys.ID)
	// TODO: find system WITH appliance details

	// TODO use credentials stored in DB (see appliance service for example
	v := libvirt.NewWithDialer(dialers.NewLocal())
	if err := v.Connect(); err != nil {
		return fmt.Errorf("cannot connect: %w", err)
	}

	// TODO find system uid
	uid := uuid.MustParse("")
	d, err := v.DomainLookupByUUID(libvirt.UUID(uid))
	if err != nil {
		return fmt.Errorf("cannot lookup %s: %w", uid.String(), err)
	}

	xmlString, err := v.DomainGetXMLDesc(d, 0)
	if err != nil {
		return fmt.Errorf("cannot get domain: %w", err)
	}
	domain := libvirtxml.Domain{}
	if err := xml.Unmarshal([]byte(xmlString), &domain); err != nil {
		return fmt.Errorf("cannot unmarshal domain XML: %w", err)
	}
	domain.OS.BootDevices = []libvirtxml.DomainBootDevice{{Dev: "network"}}
	bytes, err := xml.Marshal(domain)
	if err != nil {
		return fmt.Errorf("cannot marshal domain XML: %w", err)
	}

	d, err = v.DomainDefineXML(string(bytes))
	if err != nil {
		return fmt.Errorf("cannot redefine domain: %w", err)
	}
	state, _, err := v.DomainGetState(d, 0)
	if err != nil {
		return fmt.Errorf("cannot get domain state: %w", err)
	}

	if state == 1 {
		// domain is running
		err = v.DomainReset(d, 0)
		if err != nil {
			return fmt.Errorf("cannot reset domain: %w", err)
		}
	} else {
		// domain was not running
		err = v.DomainCreate(d)
		if err != nil {
			return fmt.Errorf("cannot create domain: %w", err)
		}
	}

	if err := v.Disconnect(); err != nil {
		return fmt.Errorf("cannot connect to libvirt: %w", err)
	}

	return nil
}
