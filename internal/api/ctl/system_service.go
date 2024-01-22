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
	"net"
	"sort"
	"strings"
	"time"

	"golang.org/x/exp/slog"
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
		facts.List = append(facts.List, existingSystem.Facts.List...)
	}

	sys = &model.System{
		Name:    system.Name,
		HwAddrs: hwAddrs.Unique(),
		Facts:   facts,
		UID:     system.UID,
	}

	if system.ApplianceName != nil {
		ad := db.GetApplianceDao(ctx)
		app, err := ad.Find(ctx, *system.ApplianceName)
		if err != nil {
			return fmt.Errorf("cannot find appliance named '%v': %w", system.ApplianceName, err)
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
		ID:      result.System.ID,
		Name:    result.System.Name,
		HwAddrs: hwa,
		Facts:   result.System.Facts.FactsMap(),
		Comment: result.System.Comment,
		UID:     result.System.UID,
	}

	payload.Appliance = &Appliance{
		ID:   result.Appliance.ID,
		Name: result.Appliance.Name,
		Kind: int16(result.Appliance.Kind),
		URI:  result.Appliance.URI,
	}

	return payload, nil
}

func (i SystemServiceImpl) Rename(ctx context.Context, pattern, newName string) error {
	dao := db.GetSystemDao(ctx)
	sys, err := dao.Find(ctx, pattern)
	if err != nil {
		return fmt.Errorf("cannot find system %s: %w", pattern, err)
	}

	sys.Name = newName
	err = dao.Rename(ctx, sys.ID, newName)
	if err != nil {
		return fmt.Errorf("cannot update: %w", err)
	}

	return nil
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
			ID:      item.ID,
			Name:    item.Name,
			HwAddrs: item.HwAddrStrings(),
			Facts:   item.Facts.FactsMap(),
			Comment: item.Comment,
		}
	}

	return result, nil
}

func (i SystemServiceImpl) Deploy(ctx context.Context, systemPattern string, imagePattern string, snippets []string, customSnippet string, ksOverride string, comment string, validUntil time.Time) error {
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

	err = daoSystem.Deploy(ctx, system.ID, image.ID, snippetIDs, customSnippet, ksOverride, comment, validUntil)
	if err != nil {
		return fmt.Errorf("cannot deploy: %w", err)
	}

	// TODO this should be done in the background via notification
	err = i.BootNetwork(ctx, systemPattern)
	if err != nil {
		return fmt.Errorf("cannot reset after deploy: %w", err)
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

	sort.Sort(logEntries)
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
