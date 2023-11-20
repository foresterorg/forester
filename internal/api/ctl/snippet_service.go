package ctl

import (
	"context"
	"fmt"
	"forester/internal/db"
	"forester/internal/model"
)

var _ SnippetService = SnippetServiceImpl{}

type SnippetServiceImpl struct{}

func (i SnippetServiceImpl) Create(ctx context.Context, name string, kind int16, body string) error {
	dao := db.GetSnippetDao(ctx)
	snippet := model.Snippet{
		Name: name,
		Kind: model.ParseSnippetKind(kind),
		Body: body,
	}

	err := dao.Create(ctx, &snippet)
	if err != nil {
		return fmt.Errorf("cannot create: %w", err)
	}

	return nil
}

func (i SnippetServiceImpl) Find(ctx context.Context, name string) (*Snippet, error) {
	dao := db.GetSnippetDao(ctx)
	result, err := dao.Find(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("cannot find: %w", err)
	}

	return &Snippet{
		ID:   result.ID,
		Name: result.Name,
		Kind: int16(result.Kind),
		Body: result.Body,
	}, nil
}

func (i SnippetServiceImpl) Edit(ctx context.Context, name string, body string) error {
	dao := db.GetSnippetDao(ctx)
	err := dao.EditByName(ctx, name, body)
	if err != nil {
		return fmt.Errorf("cannot find: %w", err)
	}

	return nil
}

func (i SnippetServiceImpl) List(ctx context.Context, limit int64, offset int64) ([]*Snippet, error) {
	dao := db.GetSnippetDao(ctx)
	ensureLimitNonzero(&limit)
	snippets, err := dao.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("cannot list: %w", err)
	}
	result := make([]*Snippet, len(snippets))
	for i, s := range snippets {
		result[i] = &Snippet{
			ID:   s.ID,
			Name: s.Name,
			Kind: int16(s.Kind),
			Body: s.Body,
		}
	}
	return result, nil
}

func (i SnippetServiceImpl) Delete(ctx context.Context, name string) error {
	dao := db.GetSnippetDao(ctx)
	err := dao.DeleteByName(ctx, name)
	if err != nil {
		return fmt.Errorf("cannot delete: %w", err)
	}
	return nil
}
