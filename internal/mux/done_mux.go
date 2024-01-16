package mux

import (
	"context"
	"forester/internal/db"
	"forester/internal/logging"
	"forester/internal/metal"
	"forester/internal/model"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
)

func MountDone(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypePlainText))

		r.Post("/{ID}", HandleDone)
	})
}

func HandleDone(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "ID"), 10, 64)
	if err != nil {
		slog.InfoContext(r.Context(), "installation what", "system_id", chi.URLParam(r, "ID"), "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slog.InfoContext(r.Context(), "installation done", "system_id", id)
	sDao := db.GetSystemDao(r.Context())
	systemAppliance, err := sDao.FindByIDRelated(r.Context(), id)
	if err != nil {
		slog.InfoContext(r.Context(), "system not found", "system_id", chi.URLParam(r, "ID"), "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if systemAppliance.ApplianceID == nil {
		slog.InfoContext(r.Context(), "system has no appliance associated", "system_id", chi.URLParam(r, "ID"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// boot libvirt manually
	if systemAppliance.Appliance.Kind == model.LibvirtApplianceKind {
		m := metal.ForKind(systemAppliance.Appliance.Kind)
		// cannot pass request context it will be cancelled
		bctx := logging.WithTraceId(context.Background(), logging.TraceId(r.Context()))
		go bootLocal(bctx, m, systemAppliance)
	}

	w.WriteHeader(http.StatusOK)
}

func bootLocal(ctx context.Context, m metal.Metal, s *model.SystemAppliance) {
	slog.InfoContext(ctx, "will boot system locally", "system_id", s.System.ID)
	time.Sleep(30 * time.Second)
	slog.InfoContext(ctx, "booting system locally", "system_id", s.System.ID)
	err := m.BootLocal(ctx, s)
	if err != nil {
		slog.InfoContext(ctx, "error during local boot", "system_id", s.System.ID, "error", err.Error())
	}
}
