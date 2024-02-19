package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/hashicorp/golang-lru/v2/expirable"

	"forester/internal/model"
)

func init() {
	GetInstallationDao = getInstallationDao
}

type instDao struct{}

func getInstallationDao(ctx context.Context) InstallationDao {
	return &instDao{}
}

func (dao instDao) FindValid(ctx context.Context, uuid uuid.UUID, state model.InstallState) (*model.Installation, error) {
	query := `SELECT * FROM installations WHERE uuid = $1 AND valid_until > current_timestamp AND state <= $2`

	result := &model.Installation{}
	err := pgxscan.Get(ctx, Pool, result, query, uuid, state)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	return result, nil
}

func (dao instDao) FindValidByState(ctx context.Context, systemId int64, state model.InstallState) ([]*model.Installation, error) {
	query := `SELECT * FROM installations WHERE system_id = $1 AND valid_until > current_timestamp AND state <= $2 ORDER BY id DESC`

	var result []*model.Installation
	rows, err := Pool.Query(ctx, query, systemId, state)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	err = pgxscan.ScanAll(&result, rows)
	if err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return result, nil
}

func (dao instDao) FindAnyByState(ctx context.Context, state model.InstallState) ([]*model.Installation, error) {
	query := `SELECT * FROM installations WHERE valid_until > current_timestamp AND state <= $2 ORDER BY valid_until DESC`

	var result []*model.Installation
	rows, err := Pool.Query(ctx, query, state)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	err = pgxscan.ScanAll(&result, rows)
	if err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return result, nil
}

var ErrUnknownSystem = errors.New("unknown system")

var NullMAC net.HardwareAddr

var installationCache *expirable.LRU[string, installationCacheEntry]

type installationCacheEntry struct {
	inst *model.Installation
	sys  *model.System
}

func init() {
	NullMAC, _ = net.ParseMAC("00:00:00:00:00:00")
	installationCache = expirable.NewLRU[string, installationCacheEntry](512, nil, 30*time.Second)
}

func (dao instDao) FindInstallationForMAC(ctx context.Context, givenMAC net.HardwareAddr) (*model.Installation, *model.System, error) {
	// lookup in cache
	if value, ok := installationCache.Get(givenMAC.String()); ok {
		return value.inst, value.sys, nil
	}

	sDao := GetSystemDao(ctx)
	iDao := GetInstallationDao(ctx)

	if givenMAC == nil {
		givenMAC = NullMAC
	}
	addrs := [2]net.HardwareAddr{
		givenMAC,
		NullMAC,
	}
	var err error
	var state model.InstallState
	var s *model.System
	var i *model.Installation

	for r, mac := range addrs {

		s, err = sDao.FindByMac(ctx, mac)
		if err != nil {
			slog.InfoContext(ctx, "unknown host by mac", "mac", mac.String())
			// try again with null mac
			continue
		}

		if r == 0 {
			state = model.BootingInstallState
		} else {
			state = model.AnyInstallState
		}

		is, err := iDao.FindValidByState(ctx, s.ID, state)
		if err != nil || len(is) < 1 {
			slog.InfoContext(ctx, "system has no active installation", "mac", mac.String())
			// try again with null mac
			continue
		} else {
			if len(is) > 1 {
				slog.WarnContext(ctx, "more than one installations for host", "mac", mac.String())
			} else {
				slog.InfoContext(ctx, "found installation for system", "mac", mac.String(), "system_id", s.ID)
			}
			// found it
			i = is[0]
			installationCache.Add(givenMAC.String(), installationCacheEntry{inst: i, sys: s})
			return i, s, nil
		}
	}

	return nil, nil, fmt.Errorf("%w: %s", ErrUnknownSystem, givenMAC.String())
}
