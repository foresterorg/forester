package db

import (
	"context"
	"errors"
	"forester/internal/model"

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

var GetSystemDao func(ctx context.Context) SystemDao

type SystemDao interface {
	FindByMac(ctx context.Context, system *model.System, mac string) error
}
