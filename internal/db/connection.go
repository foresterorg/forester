package db

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	_ "github.com/georgysavva/scany/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"

	"forester/internal/config"
)

// Pool is the main connection pool for the whole application
var Pool *pgxpool.Pool

func getConnString(prefix, schema string) string {
	if len(config.Database.Password) > 0 {
		return fmt.Sprintf("%s://%s:%s@%s:%d/%s?search_path=%s",
			prefix,
			url.QueryEscape(config.Database.User),
			url.QueryEscape(config.Database.Password),
			config.Database.Host,
			config.Database.Port,
			config.Database.Name,
			schema)
	} else {
		return fmt.Sprintf("%s://%s@%s:%d/%s?search_path=%s",
			prefix,
			url.QueryEscape(config.Database.User),
			config.Database.Host,
			config.Database.Port,
			config.Database.Name,
			schema)
	}
}

// Initialize creates connection pool. Close must be called when done.
func Initialize(ctx context.Context, schema string) error {
	var err error
	if schema == "" {
		schema = "public"
	}

	// register and setup logging configuration
	connStr := getConnString("postgres", schema)
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("unable to parse db configuration: %w", err)
	}
	slog.DebugContext(ctx, "connecting to database", "conn_string", connStr)

	poolConfig.MaxConns = config.Database.MaxConn
	poolConfig.MinConns = config.Database.MinConn
	poolConfig.MaxConnLifetime = config.Database.MaxLifetime
	poolConfig.MaxConnIdleTime = config.Database.MaxIdleTime

	logLevel, configErr := tracelog.LogLevelFromString(config.Database.LogLevel)
	if configErr != nil {
		return fmt.Errorf("cannot parse db log level configuration: %w", configErr)
	}
	poolConfig.ConnConfig.Tracer = NewTracerLogger(slog.Default(), logLevel)

	Pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}

	err = Pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("unable to ping the database: %w", err)
	}

	return nil
}

func Close() {
	slog.Debug("closing database")
	Pool.Close()
}
