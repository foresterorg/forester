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

func (dao imageDao) Create(ctx context.Context, image *model.Image) error {
	query := `INSERT INTO images (name, liveimg_sha256) VALUES ($1, $2) RETURNING id`

	err := Pool.QueryRow(ctx, query, image.Name, image.LiveimgSha256).Scan(&image.ID)
	if err != nil {
		return fmt.Errorf("db error: %w", err)
	}

	return nil
}

func (dao imageDao) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	query := `SELECT * FROM images WHERE id = $1 LIMIT 1`

	result := &model.Image{}
	err := pgxscan.Get(ctx, Pool, result, query, id)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	return result, nil
}

func (dao imageDao) Find(ctx context.Context, pattern string) (*model.Image, error) {
	query := `SELECT * FROM images WHERE name = $1 LIMIT 1`

	result := &model.Image{}
	err := pgxscan.Get(ctx, Pool, result, query, pattern)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	return result, nil
}

func (dao imageDao) List(ctx context.Context, limit, offset int64) ([]*model.Image, error) {
	query := `SELECT * FROM images ORDER BY id LIMIT $1 OFFSET $2`

	var result []*model.Image
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

func (dao imageDao) Update(ctx context.Context, image *model.Image) error {
	query := `UPDATE images SET name = $2, liveimg_sha256 = $3 WHERE id = $1`

	tag, err := Pool.Exec(ctx, query, image.ID, image.Name, image.LiveimgSha256)
	if err != nil {
		return fmt.Errorf("update error: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("expected 1 row: %w", ErrAffectedMismatch)
	}

	return nil
}

func (dao imageDao) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM images WHERE id = $1`

	tag, err := Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete error: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("expected 1 row: %w", ErrAffectedMismatch)
	}

	return nil
}
