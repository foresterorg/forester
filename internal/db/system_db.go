package db

import (
	"context"
	"fmt"
	"forester/internal/model"
	"time"
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

func (dao systemDao) FindByMac(ctx context.Context, mac string) (*model.System, error) {
	result := &model.System{}
	result.ImageID = 5
	result.AcquiredAt = time.Now()
	return result, nil
}
