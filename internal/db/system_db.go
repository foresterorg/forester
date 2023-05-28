package db

import (
	"context"
	"fmt"
	"forester/internal/model"
	"forester/internal/ptr"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
)

func init() {
	GetSystemDao = getSystemDao
}

type systemDao struct{}

func getSystemDao(ctx context.Context) SystemDao {
	return &systemDao{}
}

func (dao systemDao) Register(ctx context.Context, sys *model.System) error {
	query := `INSERT INTO systems (hwaddrs, facts) VALUES ($1, $2) RETURNING id`

	err := Pool.QueryRow(ctx, query, sys.HwAddrs, sys.Facts).Scan(&sys.ID)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	return nil
}

func (dao systemDao) List(ctx context.Context, limit, offset int64) ([]*model.System, error) {
	query := `SELECT * FROM systems ORDER BY id LIMIT $1 OFFSET $2`

	var result []*model.System
	rows, err := Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	err = pgxscan.ScanAll(&result, rows)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	return result, nil
}

func (dao systemDao) FindByMac(ctx context.Context, mac string) (*model.System, error) {
	result := &model.System{}
	result.ImageID = ptr.ToInt64(5)
	result.AcquiredAt = time.Now()
	return result, nil
}
