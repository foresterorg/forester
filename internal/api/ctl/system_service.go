package ctl

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/db"
	"forester/internal/logstore"
	"forester/internal/metal"
	"forester/internal/model"
	"forester/internal/mux"
	"golang.org/x/exp/slog"
	"net"
	"strings"
)

var _ SystemService = SystemServiceImpl{}

type SystemServiceImpl struct{}

func (i SystemServiceImpl) Register(ctx context.Context, system *NewSystem) error {
	var sys *model.System
	var existingSystem *model.System
	var hwAddrs model.HwAddrSlice
	var err error

	dao := db.GetSystemDao(ctx)

	for _, a := range system.HwAddrs {
		mac, err := net.ParseMAC(a)
		if err != nil {
			return fmt.Errorf("cannot parse hardware address '%s': %w", a, err)
		}
		slog.DebugContext(ctx, "searching for existing host", "mac", mac.String())
		sys, err := dao.FindByMac(ctx, mac)
		if err != nil && !errors.Is(err, db.ErrNoRows) {
			return fmt.Errorf("cannot search existing systems for mac '%s': %w", mac.String(), err)
		}
		if sys != nil {
			slog.DebugContext(ctx, "found existing host", "mac", mac.String(), "id", sys.ID)
			existingSystem = sys
		}

		hwAddrs = append(hwAddrs, mac)
	}

	var facts model.Facts
	for k, v := range system.Facts {
		facts.List = append(facts.List, model.Fact{Key: k, Value: v})
	}
	if existingSystem != nil {
		for _, fn := range existingSystem.Facts.List {
			facts.List = append(facts.List, fn)
		}
	}

	sys = &model.System{
		HwAddrs: hwAddrs.Unique(),
		Facts:   facts,
		UID:     system.UID,
	}

	if system.ApplianceName != nil {
		ad := db.GetApplianceDao(ctx)
		app, err := ad.Find(ctx, *system.ApplianceName)
		if err != nil {
			return fmt.Errorf("cannot find appliance named '%s': %w", system.ApplianceName, err)
		}
		sys.ApplianceID = &app.ID
	}

	if existingSystem != nil {
		slog.DebugContext(ctx, "updating existing system record",
			"id", existingSystem.ID,
			"mac", sys.HwAddrString(),
		)
		err = dao.RegisterExisting(ctx, existingSystem.ID, sys)
	} else {
		slog.DebugContext(ctx, "creating new system record", "mac", sys.HwAddrString())
		err = dao.Register(ctx, sys)
	}
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
	for i := range result.System.HwAddrs {
		hwa[i] = result.System.HwAddrs[i].String()
	}

	payload := &System{
		ID:          result.System.ID,
		Name:        result.System.Name,
		HwAddrs:     hwa,
		Facts:       result.System.Facts.FactsMap(),
		Acquired:    result.System.Acquired,
		AcquiredAt:  result.System.AcquiredAt,
		ImageID:     result.System.ImageID,
		Comment:     result.System.Comment,
		UID:         result.System.UID,
		InstallUUID: result.System.InstallUUID.String(),
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
	ensureLimitNonzero(&limit)
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

func (i SystemServiceImpl) Acquire(ctx context.Context, systemPattern, imagePattern, comment string, snippets []string) error {
	daoSystem := db.GetSystemDao(ctx)
	daoImage := db.GetImageDao(ctx)
	daoSnip := db.GetSnippetDao(ctx)

	image, err := daoImage.Find(ctx, imagePattern)
	if err != nil {
		return fmt.Errorf("cannot find: %w", err)
	}
	system, err := daoSystem.Find(ctx, systemPattern)
	if err != nil {
		return fmt.Errorf("cannot find: %w", err)
	}

	snippetIDs := make([]int64, len(snippets))
	for i, snippet := range snippets {
		slog.DebugContext(ctx, "checking snippet", "name", snippet)
		s, err := daoSnip.Find(ctx, snippet)
		if err != nil {
			return fmt.Errorf("cannot find snippet named %s: %w", snippet, err)
		}
		snippetIDs[i] = s.ID
	}

	err = daoSystem.Acquire(ctx, system.ID, image.ID, comment, snippetIDs)
	if err != nil {
		return fmt.Errorf("cannot acquire: %w", err)
	}

	err = i.BootNetwork(ctx, systemPattern)
	if err != nil {
		return fmt.Errorf("cannot reset after acquire: %w", err)
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

func (i SystemServiceImpl) BootNetwork(ctx context.Context, systemPattern string) error {
	dao := db.GetSystemDao(ctx)
	system, err := dao.FindRelated(ctx, systemPattern)
	if err != nil {
		return fmt.Errorf("cannot find: %w", err)
	}

	return metal.BootNetwork(ctx, system)
}

func (i SystemServiceImpl) BootLocal(ctx context.Context, systemPattern string) error {
	dao := db.GetSystemDao(ctx)
	system, err := dao.FindRelated(ctx, systemPattern)
	if err != nil {
		return fmt.Errorf("cannot find: %w", err)
	}

	return metal.BootLocal(ctx, system)
}

func (i SystemServiceImpl) Kickstart(ctx context.Context, pattern string) (string, error) {
	dao := db.GetSystemDao(ctx)
	system, err := dao.Find(ctx, pattern)
	if err != nil {
		return "", err
	}

	buf := strings.Builder{}
	err = mux.RenderKickstartForSystem(ctx, system, &buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (i SystemServiceImpl) Logs(ctx context.Context, systemPattern string) ([]*LogEntry, error) {
	dao := db.GetSystemDao(ctx)
	system, err := dao.Find(ctx, systemPattern)
	if err != nil {
		return nil, err
	}

	logEntries, err := logstore.LogsForSystem(ctx, system.ID)
	if err != nil {
		return nil, err
	}

	result := make([]*LogEntry, 0, len(logEntries))
	for _, le := range logEntries {
		nle := LogEntry{
			Path:       le.Path,
			Size:       le.Size,
			CreatedAt:  le.CreatedAt,
			ModifiedAt: le.ModifiedAt,
		}
		result = append(result, &nle)
	}

	return result, nil
}
