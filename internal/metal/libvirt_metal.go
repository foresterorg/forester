package metal

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"forester/internal/db"
	"forester/internal/model"
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/digitalocean/go-libvirt"
	"github.com/digitalocean/go-libvirt/socket"
	"github.com/digitalocean/go-libvirt/socket/dialers"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"libvirt.org/go/libvirtxml"
)

type LibvirtMetal struct{}

func updateDomainBootDeviceXML(xmlString, device string) (string, error) {
	domain := libvirtxml.Domain{}
	if err := xml.Unmarshal([]byte(xmlString), &domain); err != nil {
		return "", fmt.Errorf("cannot unmarshal domain XML: %w", err)
	}
	domain.OS.BootDevices = []libvirtxml.DomainBootDevice{{Dev: device}}
	bytes, err := xml.Marshal(domain)
	if err != nil {
		return "", fmt.Errorf("cannot marshal domain XML: %w", err)
	}

	return string(bytes), nil
}

var ErrUnsupportedLibvirtScheme = errors.New("unsupported scheme, valid options are: unix")

func libvirtFromURI(ctx context.Context, uri string) (*libvirt.Libvirt, error) {
	var dialer socket.Dialer
	var v *libvirt.Libvirt

	slog.DebugCtx(ctx, "connecting to libvirt", "uri", uri)
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("cannot parse: %w", err)
	}

	if parsed.Scheme == "unix" {
		dialer = dialers.NewLocal(dialers.WithSocket(parsed.Path))
		v = libvirt.NewWithDialer(dialer)
	} else if parsed.Scheme == "tcp" {
		host, _, _ := net.SplitHostPort(parsed.Host)
		dialer = dialers.NewRemote(host, dialers.UsePort(parsed.Port()))
		slog.DebugCtx(ctx, "dialer", "d", dialer)
		v = libvirt.NewWithDialer(dialer)
	} else {
		return nil, ErrUnsupportedLibvirtScheme
	}

	return v, nil
}

func bootDevice(ctx context.Context, system *model.SystemAppliance, device string) error {
	daoApp := db.GetApplianceDao(ctx)
	app, err := daoApp.FindByID(ctx, *system.ApplianceID)
	if err != nil {
		return fmt.Errorf("cannot find appliance with id %d: %w", system.ApplianceID, err)
	}

	v, err := libvirtFromURI(ctx, app.URI)
	if err != nil {
		return fmt.Errorf("URI '%s' error: %w", app.URI, err)
	}
	if err := v.Connect(); err != nil {
		return fmt.Errorf("cannot connect: %w", err)
	}
	defer v.Disconnect()

	uid := uuid.MustParse(*system.UID)
	d, err := v.DomainLookupByUUID(libvirt.UUID(uid))
	if err != nil {
		return fmt.Errorf("cannot lookup %s: %w", uid.String(), err)
	}

	xmlString, err := v.DomainGetXMLDesc(d, 0)
	if err != nil {
		return fmt.Errorf("cannot get domain: %w", err)
	}

	newXML, err := updateDomainBootDeviceXML(xmlString, device)

	if err != nil {
		return fmt.Errorf("cannot update domain XML: %w", err)
	}

	d, err = v.DomainDefineXML(newXML)
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

	return nil
}

func (m LibvirtMetal) Enlist(ctx context.Context, app *model.Appliance, pattern string) ([]*EnlistResult, error) {
	v, err := libvirtFromURI(ctx, app.URI)
	if err != nil {
		return nil, fmt.Errorf("URI '%s' error: %w", app.URI, err)
	}

	if err := v.Connect(); err != nil {
		return nil, fmt.Errorf("cannot connect: %w", err)
	}
	defer v.Disconnect()

	domains, _, err := v.ConnectListAllDomains(1, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot list domains: %w", err)
	}

	rg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("cannot compile regular expression '%s': %w", pattern, err)
	}

	var result []*EnlistResult
	for _, d := range domains {
		uid := uuid.UUID(d.UUID).String()
		xmlString, err := v.DomainGetXMLDesc(d, 0)
		if err != nil {
			return nil, fmt.Errorf("cannot get domain details: %w", err)
		}
		domain := libvirtxml.Domain{}
		if err := xml.Unmarshal([]byte(xmlString), &domain); err != nil {
			return nil, fmt.Errorf("cannot unmarshal domain XML: %w", err)
		}
		if rg.MatchString(domain.Name) {
			var addrs []string
			for _, iface := range domain.Devices.Interfaces {
				addrs = append(addrs, iface.MAC.Address)
			}

			facts := map[string]string{
				"vm_title":      domain.Title,
				"vm_emulator":   domain.Devices.Emulator,
				"vm_bootloader": domain.Bootloader,
			}

			er := &EnlistResult{
				HwAddrs: addrs,
				Facts:   facts,
				UID:     uid,
			}
			slog.InfoCtx(ctx, "found system",
				"mac", strings.Join(addrs, ","),
				"uuid", uid,
				"appliance", app.Name,
				"name", d.Name,
			)
			result = append(result, er)
		} else {
			slog.InfoCtx(ctx, "system does not match the pattern",
				"pattern", pattern,
				"appliance", app.Name,
				"name", d.Name,
				"uuid", uid,
			)
		}
	}

	return result, nil
}

func (m LibvirtMetal) BootNetwork(ctx context.Context, system *model.SystemAppliance) error {
	return bootDevice(ctx, system, "network")
}

func (m LibvirtMetal) BootLocal(ctx context.Context, system *model.SystemAppliance) error {
	return bootDevice(ctx, system, "hd")
}
