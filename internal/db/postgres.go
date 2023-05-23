package db

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/model"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

var (
	// ErrNoRows is returned when there are no rows in the result
	// Typically, REST requests should end up with 404 error
	ErrNoRows = pgx.ErrNoRows

	// ErrAffectedMismatch is returned when unexpected number of affected rows
	// was returned for INSERT, UPDATE and DELETE queries.
	// Typically, REST requests should end up with 409 error
	ErrAffectedMismatch = errors.New("unexpected affected rows")
)

var GetImageDao func(ctx context.Context) ImageDao

type ImageDao interface {
	Create(ctx context.Context, img *model.Image) error
	GetById(ctx context.Context, img *model.Image, id int64) error
	List(ctx context.Context, img *[]model.Image, limit, offset int64) error
	Delete(ctx context.Context, id int64) error
}

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
