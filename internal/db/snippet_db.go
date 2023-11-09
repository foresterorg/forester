package db

import (
	"context"
	"fmt"
	"forester/internal/model"

	"github.com/georgysavva/scany/v2/pgxscan"
)

func init() {
	GetSnippetDao = getSnippetDao
}

type snippetDao struct{}

func getSnippetDao(_ context.Context) SnippetDao {
	return &snippetDao{}
}

func (dao snippetDao) Create(ctx context.Context, a *model.Snippet) error {
	query := `INSERT INTO snippets (name, kind, body) VALUES ($1, $2, $3) RETURNING id`

	err := Pool.QueryRow(ctx, query, a.Name, a.Kind, a.Body).Scan(&a.ID)
	if err != nil {
		return fmt.Errorf("insert error: %w", err)
	}

	return nil
}

func (dao snippetDao) Find(ctx context.Context, name string) (*model.Snippet, error) {
	query := `SELECT * FROM snippets WHERE name = $1 LIMIT 1`

	result := &model.Snippet{}
	err := pgxscan.Get(ctx, Pool, result, query, name)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	return result, nil
}

func (dao snippetDao) FindByID(ctx context.Context, id int64) (*model.Snippet, error) {
	query := `SELECT * FROM snippets WHERE id = $1 LIMIT 1`

	result := &model.Snippet{}
	err := pgxscan.Get(ctx, Pool, result, query, id)
	if err != nil {
		return nil, fmt.Errorf("select error: %w", err)
	}

	return result, nil
}

func (dao snippetDao) List(ctx context.Context, limit, offset int64) ([]*model.Snippet, error) {
	query := `SELECT * FROM snippets ORDER BY id LIMIT $1 OFFSET $2`

	var result []*model.Snippet
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

func (dao snippetDao) EditByName(ctx context.Context, name, body string) error {
	query := `UPDATE snippets SET body = $2 WHERE name = $1`

	tag, err := Pool.Exec(ctx, query, name, body)
	if err != nil {
		return fmt.Errorf("update error: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("expected 1 row: %w", ErrAffectedMismatch)
	}

	return nil
}

func (dao snippetDao) DeleteByName(ctx context.Context, name string) error {
	query := `DELETE FROM snippets WHERE name = $1`

	tag, err := Pool.Exec(ctx, query, name)
	if err != nil {
		return fmt.Errorf("delete error: %w", err)
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("expected 1 row: %w", ErrAffectedMismatch)
	}

	return nil
}
