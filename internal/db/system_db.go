package db

import (
	"context"
	"fmt"
	"forester/internal/model"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/slog"
	"net"
	"strconv"
	"time"
)

func init() {
	GetSystemDao = getSystemDao
}

type systemDao struct{}

func getSystemDao(ctx context.Context) SystemDao {
	return &systemDao{}
}

func (dao systemDao) Register(ctx context.Context, sys *model.System) error {
	if sys.Name == "" {
		query := `INSERT INTO systems (hwaddrs, facts, appliance_id, uid) VALUES ($1, $2, $3, $4) RETURNING id`

		err := Pool.QueryRow(ctx, query, sys.HwAddrs, sys.Facts, sys.ApplianceID, sys.UID).Scan(&sys.ID)
		if err != nil {
			return fmt.Errorf("insert error: %w", err)
		}
	} else {
		query := `INSERT INTO systems (hwaddrs, facts, appliance_id, uid, name) VALUES ($1, $2, $3, $4, $5) RETURNING id`

		err := Pool.QueryRow(ctx, query, sys.HwAddrs, sys.Facts, sys.ApplianceID, sys.UID, sys.Name).Scan(&sys.ID)
		if err != nil {
			return fmt.Errorf("insert error: %w", err)
		}
	}

	return nil
}

func (dao systemDao) RegisterExisting(ctx context.Context, id int64, sys *model.System) error {
	query := `UPDATE systems SET hwaddrs = $2, facts = $3 WHERE id = $1 RETURNING id`

	err := Pool.QueryRow(ctx, query, id, sys.HwAddrs, sys.Facts).Scan(&sys.ID)
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

func (dao systemDao) Acquire(ctx context.Context, systemId, imageId int64, force bool, snippets []int64, snippetText, ksOverride, comment string, validUntil time.Time) error {
	txErr := WithTransaction(ctx, func(tx pgx.Tx) error {
		insertQuery := `INSERT INTO installations (system_id, image_id, snippet_text, kickstart_override, comment, valid_until) VALUES
			($1, $2, $3, $4, $5, $6) RETURNING id`
		var instID int64

		err := Pool.QueryRow(ctx, insertQuery, systemId, imageId, snippetText, ksOverride, comment, validUntil).Scan(&instID)
		if err != nil {
			return fmt.Errorf("installation insert error: %w", err)
		}

		updateQuery := `UPDATE systems SET
		acquired = true,
		acquired_at = current_timestamp
		WHERE id = $1`

		if !force {
			updateQuery += " AND acquired = false"
		}

		tag, err := tx.Exec(ctx, updateQuery, systemId)
		if err != nil {
			return fmt.Errorf("update error: %w", err)
		}

		if tag.RowsAffected() != 1 {
			return fmt.Errorf("cannot find unacquired system with ID=%d: %w", systemId, ErrAffectedMismatch)
		}

		deleteQuery := `DELETE FROM installations_snippets WHERE installation_id = $1`
		tag, err = tx.Exec(ctx, deleteQuery, instID)
		if err != nil {
			return fmt.Errorf("delete snippets error: %w", err)
		}
		slog.DebugContext(ctx, "deleted existing snippets", "affected", tag.RowsAffected())

		batch := &pgx.Batch{}
		for _, s := range snippets {
			batch.Queue("INSERT INTO installations_snippets VALUES ($1, $2)", instID, s)
		}
		br := tx.SendBatch(ctx, batch)
		defer br.Close()
		tag, err = br.Exec()

		if err != nil {
			return fmt.Errorf("batch insert error: %w", err)
		}

		if tag.RowsAffected() != 1 {
			return fmt.Errorf("batch insert row mismatch, expected %d got %d", len(snippets), tag.RowsAffected())
		}
		slog.DebugContext(ctx, "saved snippets", "affected", tag.RowsAffected())

		return nil
	})

	return txErr
}

func (dao systemDao) Rename(ctx context.Context, systemId int64, newName string) error {
	query := `UPDATE systems SET
		name = $2
		WHERE id = $1`

	tag, err := Pool.Exec(ctx, query, systemId, newName)
	if err != nil {
		return fmt.Errorf("update error: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("cannot find acquired system with ID=%d: %w", systemId, ErrAffectedMismatch)
	}

	return nil
}

func (dao systemDao) Release(ctx context.Context, systemId int64) error {
	query := `UPDATE systems SET
		acquired = false,
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

func (dao systemDao) FindRelated(ctx context.Context, pattern string) (*model.SystemAppliance, error) {
	if mac, err := net.ParseMAC(pattern); err == nil {
		return dao.FindByMacRelated(ctx, mac)
	}

	name := fmt.Sprintf("%%%s%%", pattern)
	result := &model.SystemAppliance{}
	query := `SELECT s.id AS "s.id",
		s.name AS "s.name",
		s.appliance_id AS "s.appliance_id",
		s.uid AS "s.uid",
		s.hwaddrs AS "s.hwaddrs",
		s.facts AS "s.facts",
		s.acquired AS "s.acquired",
		s.acquired_at AS "s.acquired_at",
		s.comment AS "s.comment",
		COALESCE(a.name, '') AS "a.name",
		COALESCE(a.kind, 0) AS "a.kind",
		COALESCE(a.uri, '') AS "a.uri"
		FROM systems AS s LEFT JOIN appliances AS a ON a.id = s.appliance_id WHERE s.name ILIKE $1 LIMIT 1`

	err := pgxscan.Get(ctx, Pool, result, query, name)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	return result, nil
}

func (dao systemDao) Find(ctx context.Context, pattern string) (*model.System, error) {
	if mac, err := net.ParseMAC(pattern); err == nil {
		return dao.FindByMac(ctx, mac)
	}
	if id, err := strconv.Atoi(pattern); err == nil {
		return dao.FindByID(ctx, int64(id))
	}

	name := fmt.Sprintf("%%%s%%", pattern)
	query := `SELECT * FROM systems WHERE name ILIKE $1 LIMIT 1`

	result := &model.System{}
	err := pgxscan.Get(ctx, Pool, result, query, name)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	return result, nil
}

func (dao systemDao) FindByMacRelated(ctx context.Context, mac net.HardwareAddr) (*model.SystemAppliance, error) {
	result := &model.SystemAppliance{}
	query := `SELECT s.id AS "s.id",
		s.name AS "s.name",
		s.appliance_id AS "s.appliance_id",
		s.uid AS "s.uid",
		s.hwaddrs AS "s.hwaddrs",
		s.facts AS "s.facts",
		s.acquired AS "s.acquired",
		s.acquired_at AS "s.acquired_at",
		s.comment AS "s.comment",
		COALESCE(a.name, '') AS "a.name",
		COALESCE(a.kind, 0) AS "a.kind",
		COALESCE(a.uri, '') AS "a.uri"
		FROM systems AS s LEFT JOIN appliances AS a ON a.id = s.appliance_id WHERE $1 = ANY(s.hwaddrs) LIMIT 1`
	err := pgxscan.Get(ctx, Pool, result, query, mac)
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

func (dao systemDao) FindByIDRelated(ctx context.Context, id int64) (*model.SystemAppliance, error) {
	result := &model.SystemAppliance{}
	query := `SELECT s.id AS "s.id",
		s.name AS "s.name",
		s.appliance_id AS "s.appliance_id",
		s.uid AS "s.uid",
		s.hwaddrs AS "s.hwaddrs",
		s.facts AS "s.facts",
		s.acquired AS "s.acquired",
		s.acquired_at AS "s.acquired_at",
		s.image_id AS "s.image_id",
		s.comment AS "s.comment",
		s.install_uuid AS "s.install_uuid",
		COALESCE(a.name, '') AS "a.name",
		COALESCE(a.kind, 0) AS "a.kind",
		COALESCE(a.uri, '') AS "a.uri"
		FROM systems AS s LEFT JOIN appliances AS a ON a.id = s.appliance_id WHERE s.id = $1 LIMIT 1`
	err := pgxscan.Get(ctx, Pool, result, query, id)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	return result, nil
}

func (dao systemDao) FindByID(ctx context.Context, id int64) (*model.System, error) {
	query := `SELECT * FROM systems WHERE id = $1 LIMIT 1`

	result := &model.System{}
	err := pgxscan.Get(ctx, Pool, result, query, id)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	return result, nil
}
