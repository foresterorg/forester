package migrations

import "embed"

//go:embed *.sql
var EmbeddedSQLMigrations embed.FS
