package metal

import (
	"context"
	"forester/internal/model"

	"golang.org/x/exp/slog"
)

type NoopMetal struct{}

func (m NoopMetal) Enlist(ctx context.Context, app *model.Appliance, pattern string) ([]*EnlistResult, error) {
	slog.InfoContext(ctx, "noop operation", "function", "Enlist")
	return []*EnlistResult{}, nil
}

func (m NoopMetal) BootNetwork(ctx context.Context, system *model.SystemAppliance) error {
	slog.InfoContext(ctx, "noop operation", "function", "BootNetwork")
	return nil
}

func (m NoopMetal) BootLocal(ctx context.Context, system *model.SystemAppliance) error {
	slog.InfoContext(ctx, "noop operation", "function", "BootLocal")
	return nil
}
