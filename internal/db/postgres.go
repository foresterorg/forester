package db

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	pgx "github.com/jackc/pgx/v5"

	"forester/internal/model"
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
	Create(ctx context.Context, image *model.Image) error
	FindByID(ctx context.Context, id int64) (*model.Image, error)
	Find(ctx context.Context, pattern string) (*model.Image, error)
	List(ctx context.Context, limit, offset int64) ([]*model.Image, error)
	Update(ctx context.Context, image *model.Image) error
	Delete(ctx context.Context, id int64) error
}

var GetSystemDao func(ctx context.Context) SystemDao

type SystemDao interface {
	Register(ctx context.Context, sys *model.System) error
	RegisterExisting(ctx context.Context, id int64, sys *model.System) error
	List(ctx context.Context, limit, offset int64) ([]*model.System, error)
	Rename(ctx context.Context, systemId int64, newName string) error
	Deploy(ctx context.Context, systemId, imageId int64, snippets []int64, snippetText, ksOverride, comment string, validUntil time.Time) error
	Find(ctx context.Context, pattern string) (*model.System, error)
	FindByID(ctx context.Context, id int64) (*model.System, error)
	FindByMac(ctx context.Context, mac net.HardwareAddr) (*model.System, error)
	FindRelated(ctx context.Context, pattern string) (*model.SystemAppliance, error)
	FindByIDRelated(ctx context.Context, id int64) (*model.SystemAppliance, error)
	FindByMacRelated(ctx context.Context, mac net.HardwareAddr) (*model.SystemAppliance, error)
}

var GetInstallationDao func(ctx context.Context) InstallationDao

type InstallationDao interface {
	FindValid(ctx context.Context, uuid uuid.UUID, state model.InstallState) (*model.Installation, error)
	FindValidByState(ctx context.Context, systemId int64, state model.InstallState) ([]*model.Installation, error)
	FindAnyByState(ctx context.Context, state model.InstallState) ([]*model.Installation, error)
	FindInstallationForMAC(ctx context.Context, givenMAC net.HardwareAddr) (*model.Installation, *model.System, error)
}

var GetApplianceDao func(ctx context.Context) ApplianceDao

type ApplianceDao interface {
	Create(ctx context.Context, a *model.Appliance) error
	Find(ctx context.Context, name string) (*model.Appliance, error)
	FindByID(ctx context.Context, id int64) (*model.Appliance, error)
	List(ctx context.Context, limit, offset int64) ([]*model.Appliance, error)
	Delete(ctx context.Context, id int64) error
}

var GetSnippetDao func(ctx context.Context) SnippetDao

type SnippetDao interface {
	Create(ctx context.Context, a *model.Snippet) error
	Find(ctx context.Context, name string) (*model.Snippet, error)
	FindByID(ctx context.Context, id int64) (*model.Snippet, error)
	FindByInstallation(ctx context.Context, instID int64) ([]model.Snippet, error)
	List(ctx context.Context, limit, offset int64) ([]*model.Snippet, error)
	EditByName(ctx context.Context, name, body string) error
	DeleteByName(ctx context.Context, name string) error
}
