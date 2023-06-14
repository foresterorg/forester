package ctl

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"forester/internal/db"
	"forester/internal/model"
	"net/url"
	"regexp"

	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"libvirt.org/go/libvirtxml"
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

var ErrUnsupportedLibvirtScheme = errors.New("unsupported scheme, valid options are: unix")

func dialerFromURI(ctx context.Context, uri string) (socket.Dialer, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("cannot parse: %w", err)
	}

	if parsed.Scheme != "unix" {
		return nil, ErrUnsupportedLibvirtScheme
	}

	v := dialers.NewLocal(dialers.WithSocket(parsed.Path))
	slog.DebugCtx(ctx, "connecting to libvirt", "socket", parsed.Path)
	return v, nil
}

func (i ApplianceServiceImpl) Enlist(ctx context.Context, name string, namePattern string) error {
	dao := db.GetApplianceDao(ctx)
	app, err := dao.Find(ctx, name)
	if err != nil {
		return fmt.Errorf("unknown appliance %s: %w", name, err)
	}

	dialer, err := dialerFromURI(ctx, app.URI)
	if err != nil {
		return fmt.Errorf("URI '%s' error: %w", app.URI, err)
	}
	v := libvirt.NewWithDialer(dialer)

	if err := v.Connect(); err != nil {
		return fmt.Errorf("cannot connect: %w", err)
	}
	defer v.Disconnect()

	domains, _, err := v.ConnectListAllDomains(1, 0)
	if err != nil {
		return fmt.Errorf("cannot list domains: %w", err)
	}

	rg, err := regexp.Compile(namePattern)
	if err != nil {
		return fmt.Errorf("cannot compile regular expression '%s': %w", namePattern, err)
	}

	for _, d := range domains {
		uid := uuid.UUID(d.UUID)
		slog.InfoCtx(ctx, "enlisting system", "name", d.Name, "uuid", uid.String())
		xmlString, err := v.DomainGetXMLDesc(d, 0)
		if err != nil {
			return fmt.Errorf("cannot get domain details: %w", err)
		}
		domain := libvirtxml.Domain{}
		if err := xml.Unmarshal([]byte(xmlString), &domain); err != nil {
			return fmt.Errorf("cannot unmarshal domain XML: %w", err)
		}
		if rg.MatchString(domain.Name) {
			for _, iface := range domain.Devices.Interfaces {
				slog.InfoCtx(ctx, "found network device", "mac", iface.MAC.Address, "uuid", uid.String())
			}
		} else {
			slog.DebugCtx(ctx, "system does not match the pattern", "pattern", namePattern, "name", d.Name, "uuid", uid.String())
		}
	}

	return nil
}

func (i ApplianceServiceImpl) Delete(ctx context.Context, name string) error {
	panic("implement me")
}
