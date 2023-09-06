package db

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"forester/internal/db/migrations"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/tern/v2/migrate"
	"golang.org/x/exp/slog"
	"io/fs"
)

type EmbeddedFS struct {
	efs *embed.FS
}

//nolint:wrapcheck
func (efs *EmbeddedFS) Open(name string) (fs.File, error) {
	return efs.efs.Open(name)
}

func NewEmbeddedFS(fs *embed.FS) *EmbeddedFS {
	return &EmbeddedFS{efs: fs}
}

func (efs *EmbeddedFS) ReadDir(dirname string) ([]fs.FileInfo, error) {
	dirEntries, err := efs.efs.ReadDir(dirname)
	if err != nil {
		return nil, fmt.Errorf("unable to read dir: %w", err)
	}
	result := make([]fs.FileInfo, 0, len(dirEntries))
	for _, de := range dirEntries {
		fi, err := de.Info()
		if err != nil {
			return nil, fmt.Errorf("unable to read dir: %w", err)
		}
		result = append(result, fi)
	}
	return result, nil
}

//nolint:wrapcheck
func (efs *EmbeddedFS) ReadFile(filename string) ([]byte, error) {
	return efs.efs.ReadFile(filename)
}

//nolint:wrapcheck
func (efs *EmbeddedFS) Glob(pattern string) (matches []string, err error) {
	return fs.Glob(efs.efs, pattern)
}

var (
	ErrNoMigrationsFound = errors.New("no migrations found")
	ErrMigration         = errors.New("unable to perform migration")
)

func Migrate(ctx context.Context, schema string) error {
	slog.DebugCtx(ctx, "checking migrations")
	if schema == "" {
		schema = "public"
	}

	conn, connErr := Pool.Acquire(ctx)
	if connErr != nil {
		return fmt.Errorf("error acquiring connection from the pool: %w", connErr)
	}
	defer conn.Release()

	mfs := NewEmbeddedFS(&migrations.EmbeddedSQLMigrations)
	table := fmt.Sprintf("%s.schema_version", schema)
	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), table)
	if err != nil {
		return fmt.Errorf("error initializing migrator: %w", err)
	}
	err = migrator.LoadMigrations(mfs)
	if err != nil {
		return fmt.Errorf("error loading migrations: %w", err)
	}
	if len(migrator.Migrations) == 0 {
		return ErrNoMigrationsFound
	}

	migrator.OnStart = func(sequence int32, name, direction, sql string) {
		slog.InfoCtx(ctx, "executing  migration", "name", name, "direction", direction)
	}

	err = migrator.Migrate(ctx)
	if err != nil {
		var mgErr *migrate.MigrationPgError
		var pgErr *pgconn.PgError
		if errors.As(err, &mgErr) && errors.As(err, &pgErr) {
			slog.ErrorCtx(ctx, "migration error", "file", pgErr.File, "code", pgErr.Code, "detail", pgErr.Detail)
			return fmt.Errorf("%w: %s", ErrMigration, pgErr.Message)
		} else {
			return fmt.Errorf("unable to perform migration: %w", err)
		}
	}

	return nil
}
