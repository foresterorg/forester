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

func updateDomainBootDeviceXML(ctx context.Context, xmlString, device string) (string, error) {
	domain := libvirtxml.Domain{}
	if err := xml.Unmarshal([]byte(xmlString), &domain); err != nil {
		return "", fmt.Errorf("cannot unmarshal domain XML: %w", err)
	}
	// clear existing Boot elements which cannot be mixed with BootDevices
	for i := range domain.Devices.Disks {
		domain.Devices.Disks[i].Boot = nil
	}
	for i := range domain.Devices.Interfaces {
		domain.Devices.Interfaces[i].Boot = nil
	}

	domain.OS.BootDevices = []libvirtxml.DomainBootDevice{{Dev: device}}
	bytes, err := xml.Marshal(domain)
	if err != nil {
		return "", fmt.Errorf("cannot marshal domain XML: %w", err)
	}

	//slog.DebugContext(ctx, "domain xml", "body", domain, "bytes", bytes)
	return string(bytes), nil
}

var ErrUnsupportedLibvirtScheme = errors.New("unsupported libvirt URI scheme")

func libvirtFromURI(ctx context.Context, uri string) (*libvirt.Libvirt, error) {
	var dialer socket.Dialer
	var v *libvirt.Libvirt

	slog.DebugContext(ctx, "connecting to libvirt", "uri", uri)
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("cannot parse: %w", err)
	}

	if parsed.Scheme == "qemu" {
		v = libvirt.NewWithDialer(dialers.NewLocal())
	} else if parsed.Scheme == "unix" {
		dialer = dialers.NewLocal(dialers.WithSocket(parsed.Path))
		v = libvirt.NewWithDialer(dialer)
	} else if parsed.Scheme == "tcp" {
		host, _, _ := net.SplitHostPort(parsed.Host)
		dialer = dialers.NewRemote(host, dialers.UsePort(parsed.Port()))
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

	newXML, err := updateDomainBootDeviceXML(ctx, xmlString, device)

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
		slog.InfoContext(ctx, "force resetting domain", "name", d.Name, "uuid", d.UUID)
		err = v.DomainReset(d, 0)
		if err != nil {
			return fmt.Errorf("cannot reset domain: %w", err)
		}
	} else {
		// domain was not running
		slog.InfoContext(ctx, "creating domain", "name", d.Name, "uuid", d.UUID)
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
		slog.ErrorContext(ctx, "cannot connect to libvirt", "err", err.Error())
		return nil, fmt.Errorf("cannot connect: %w", err)
	}
	defer v.Disconnect()

	domains, ret, err := v.ConnectListAllDomains(1, 0)
	if err != nil {
		slog.ErrorContext(ctx, "cannot list libvirt domains", "err", err.Error())
		return nil, fmt.Errorf("cannot list domains: %w", err)
	}
	slog.DebugContext(ctx, "listed all libvirt domains", "count", len(domains), "return", ret)

	rg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("cannot compile regular expression '%s': %w", pattern, err)
	}

	var result []*EnlistResult
	for _, d := range domains {
		uid := uuid.UUID(d.UUID).String()
		slog.DebugContext(ctx, "looking up vm", "uuid", uid)
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
			slog.InfoContext(ctx, "found system",
				"mac", strings.Join(addrs, ","),
				"uuid", uid,
				"appliance", app.Name,
				"name", d.Name,
			)
			result = append(result, er)
		} else {
			slog.InfoContext(ctx, "system does not match the pattern",
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
