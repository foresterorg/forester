package db

import (
	"context"
	"fmt"
	"forester/internal/model"

	"github.com/georgysavva/scany/v2/pgxscan"
)

func init() {
	GetImageDao = getImageDao
}

type imageDao struct{}

func getImageDao(ctx context.Context) ImageDao {
	return &imageDao{}
}

func (dao imageDao) Create(ctx context.Context, img *model.Image) error {
	query := `INSERT INTO images (name) VALUES ($1) RETURNING id`

	err := Pool.QueryRow(ctx, query, img.Name).Scan(&img.ID)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	return nil
}

func (dao imageDao) GetById(ctx context.Context, img *model.Image, id int64) error {
	query := `SELECT * FROM images WHERE id = $1 LIMIT 1`

	err := pgxscan.Get(ctx, Pool, img, query, id)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	return nil
}

func (dao imageDao) List(ctx context.Context, img *[]model.Image, limit, offset int64) error {
	query := `SELECT * FROM images ORDER BY id LIMIT $1 OFFSET $2`

	rows, err := Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	err = pgxscan.ScanAll(img, rows)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	return nil
}

func (dao imageDao) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM images WHERE id = $1`

	tag, err := Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("expected 1 row: %w", ErrAffectedMismatch)
	}

	return nil
}
