package db

import (
	"context"
	"fmt"
	"forester/internal/model"
	"net"
	"strings"

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
		return fmt.Errorf("insert error: %w", err)
	}

	return nil
}

func (dao systemDao) List(ctx context.Context, limit, offset int64) ([]*model.System, error) {
	query := `SELECT * FROM systems ORDER BY id LIMIT $1 OFFSET $2`

	var result []*model.System
	rows, err := Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	err = pgxscan.ScanAll(&result, rows)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	return result, nil
}

func (dao systemDao) Acquire(ctx context.Context, systemId, imageId int64, comment string) error {
	query := `UPDATE systems SET
		acquired = true,
		acquired_at = current_timestamp,
		image_id = $2,
		comment = $3
		WHERE id = $1 AND acquired = false`

	tag, err := Pool.Exec(ctx, query, systemId, imageId, comment)
	if err != nil {
		return fmt.Errorf("update error: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("cannot find unacquired system with ID=%d: %w", systemId, ErrAffectedMismatch)
	}

	return nil
}

func (dao systemDao) Release(ctx context.Context, systemId int64) error {
	query := `UPDATE systems SET
		acquired = false,
		image_id = NULL,
		comment = ''
		WHERE id = $1 AND acquired = true`

	tag, err := Pool.Exec(ctx, query, systemId)
	if err != nil {
		return fmt.Errorf("update error: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("cannot find acquired system with ID=%d: %w", systemId, ErrAffectedMismatch)
	}

	return nil
}

func (dao systemDao) Find(ctx context.Context, pattern string) (*model.System, error) {
	if mac, err := net.ParseMAC(pattern); err == nil {
		return dao.FindByMac(ctx, mac)
	}

	query := `SELECT * FROM systems WHERE name = $1 LIMIT 1`
	name := strings.Title(pattern)

	result := &model.System{}
	err := pgxscan.Get(ctx, Pool, result, query, name)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	return result, nil
}

func (dao systemDao) FindByMac(ctx context.Context, mac net.HardwareAddr) (*model.System, error) {
	query := `SELECT * FROM systems WHERE $1 = ANY(hwaddrs) LIMIT 1`

	result := &model.System{}
	err := pgxscan.Get(ctx, Pool, result, query, mac)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	return result, nil
}
