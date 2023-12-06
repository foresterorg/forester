package db

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/model"
	"net"

	"github.com/georgysavva/scany/v2/pgxscan"
	"golang.org/x/exp/slog"
)

func init() {
	GetInstallationDao = getInstallationDao
}

type instDao struct{}

func getInstallationDao(ctx context.Context) InstallationDao {
	return &instDao{}
}

func (dao instDao) FindValidByState(ctx context.Context, systemId int64, state model.InstallState) ([]*model.Installation, error) {
	query := `SELECT * FROM installations WHERE system_id = $1 AND valid_until > current_timestamp AND state <= $2 ORDER BY valid_until DESC`

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

func init() {
	NullMAC, _ = net.ParseMAC("00:00:00:00:00:00")
}

func (dao instDao) FindInstallationForMAC(ctx context.Context, givenMAC net.HardwareAddr) (*model.Installation, *model.System, error) {

	// TODO add caching with 30 seconds expiration here
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
			i = is[0]
			return i, s, nil
		}
	}

	return nil, nil, fmt.Errorf("%w: %s", ErrUnknownSystem, givenMAC.String())
}
