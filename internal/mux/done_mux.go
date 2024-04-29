package mux

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"forester/internal/db"
	"forester/internal/logging"
	"forester/internal/metal"
	"forester/internal/model"
)

func MountDone(r *chi.Mux) {
	r.Group(func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypePlainText))

		r.Post("/{UUID}", HandleDone)
	})
}

func HandleDone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := uuid.Parse(chi.URLParam(r, "UUID"))
	if err != nil {
		slog.InfoContext(ctx, "cannot parse installation UUID", "uuid", chi.URLParam(r, "UUID"), "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	iDao := db.GetInstallationDao(ctx)
	inst, err := iDao.FindValid(ctx, id, model.InstallingInstallState)
	if err != nil {
		slog.InfoContext(ctx, "installation not found", "uuid", chi.URLParam(r, "UUID"), "err", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	slog.DebugContext(ctx, "installation done - system will be started soon", "system_id", id)
	sDao := db.GetSystemDao(ctx)
	systemAppliance, err := sDao.FindByIDRelated(ctx, inst.SystemID)
	if err != nil {
		slog.InfoContext(ctx, "system not found", "system_id", inst.SystemID, "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if systemAppliance.ApplianceID == nil {
		slog.InfoContext(ctx, "system has no appliance associated", "system_id", inst.SystemID)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// boot libvirt manually
	if systemAppliance.Appliance.Kind != model.LibvirtApplianceKind {
		slog.InfoContext(ctx, "system appliance is not libvirt", "system_id", inst.SystemID)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m := metal.ForKind(systemAppliance.Appliance.Kind)
	// cannot pass request context it will be cancelled
	bctx := logging.WithTraceId(context.Background(), logging.TraceId(ctx))
	go bootLocal(bctx, m, systemAppliance)

	w.WriteHeader(http.StatusOK)
}

func bootLocal(ctx context.Context, m metal.Metal, s *model.SystemAppliance) {
	slog.InfoContext(ctx, "scheduled system reboot", "system_id", s.System.ID)
	go func() {
		time.Sleep(6 * time.Second)
		slog.InfoContext(ctx, "booting system locally", "system_id", s.System.ID)
		err := m.BootLocal(ctx, s)
		if err != nil {
			slog.InfoContext(ctx, "error during local boot", "system_id", s.System.ID, "error", err.Error())
		}
	}()
}
