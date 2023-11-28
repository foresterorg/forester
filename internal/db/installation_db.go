package db

import (
	"context"
	"fmt"
	"forester/internal/model"
	"github.com/georgysavva/scany/v2/pgxscan"
)

func init() {
	GetInstallationDao = getInstallationDao
}

type instDao struct{}

func getInstallationDao(ctx context.Context) InstallationDao {
	return &instDao{}
}

func (dao instDao) FindValidByState(ctx context.Context, systemId int64, state model.InstallState) ([]*model.Installation, error) {
	query := `SELECT * FROM installations WHERE system_id = $1 AND valid_until > current_timestamp AND state < $2 ORDER BY valid_until DESC`

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
	query := `SELECT * FROM installations WHERE valid_until > current_timestamp AND state < $2 ORDER BY valid_until DESC`

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
