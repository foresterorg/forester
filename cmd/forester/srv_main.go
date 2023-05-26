package main

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/config"
	"forester/internal/db"
	"forester/internal/logging"
	"forester/internal/mux"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
)

func main() {
	ctx := context.Background()
	logging.Initialize(slog.LevelDebug)

	err := config.Initialize("config/forester.env")
	if err != nil {
		panic(err)
	}

	err = db.Initialize(ctx, "public")
	if err != nil {
		return
	}

	rootRouter := chi.NewRouter()
	bootRouter := chi.NewRouter()
	ksRouter := chi.NewRouter()

	rootRouter.Use(mux.TraceIdMiddleware)

	mux.MountKickstart(bootRouter)
	mux.MountKickstart(ksRouter)
	rootRouter.Mount("/boot", bootRouter)
	rootRouter.Mount("/ks", ksRouter)

	rootServer := http.Server{
		Addr:    fmt.Sprintf(":%d", config.Application.Port),
		Handler: rootRouter,
	}

	waitForSignal := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		if err := rootServer.Shutdown(context.Background()); err != nil {
			slog.ErrorCtx(ctx, "shutdown error", "err", err)
		}
		close(waitForSignal)
	}()

	slog.DebugCtx(ctx, "starting service", "port", config.Application.Port)

	if err := rootServer.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorCtx(ctx, "listen error", "err", err)
		}
	}

	<-waitForSignal

	slog.DebugCtx(ctx, "shutdown complete")
}
