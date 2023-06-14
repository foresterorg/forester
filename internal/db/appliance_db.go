package db

import (
	"context"
	"fmt"
	"forester/internal/model"

	"github.com/georgysavva/scany/v2/pgxscan"
)

func init() {
	GetApplianceDao = getApplianceDao
}

type applianceDao struct{}

func getApplianceDao(ctx context.Context) ApplianceDao {
	return &applianceDao{}
}

func (dao applianceDao) Create(ctx context.Context, a *model.Appliance) error {
	query := `INSERT INTO appliances (name, kind, uri) VALUES ($1, $2, $3) RETURNING id`

	err := Pool.QueryRow(ctx, query, a.Name, a.Kind, a.URI).Scan(&a.ID)
	if err != nil {
		return fmt.Errorf("insert error: %w", err)
	}

	return nil
}

func (dao applianceDao) List(ctx context.Context, limit, offset int64) ([]*model.Appliance, error) {
	query := `SELECT * FROM appliances ORDER BY id LIMIT $1 OFFSET $2`

	var result []*model.Appliance
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

func (dao applianceDao) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM appliances WHERE id = $1`

	tag, err := Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete error: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("expected 1 row: %w", ErrAffectedMismatch)
	}

	return nil
}
