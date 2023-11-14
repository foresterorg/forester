package main

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/api/ctl"
	"forester/internal/config"
	"forester/internal/db"
	"forester/internal/img"
	"forester/internal/logging"
	"forester/internal/logstore"
	"forester/internal/mux"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-h" {
		fmt.Printf("Environment variables:\n%s\n", config.HelpText())
		os.Exit(1)
	}
	ctx := context.Background()

	logging.Initialize(config.ParsedLoggingLevel())

	err := config.Initialize("config/forester.env")
	if err != nil {
		panic(err)
	}

	syslog, err := logstore.Start(ctx)
	defer syslog.Stop()
	if err != nil {
		panic(err)
	}

	err = db.Initialize(ctx, "public")
	if err != nil {
		return
	}

	err = db.Migrate(ctx, "public")
	if err != nil {
		return
	}

	rootRouter := chi.NewRouter()
	bootRouter := chi.NewRouter()
	imgRouter := chi.NewRouter()
	ksRouter := chi.NewRouter()
	doneRouter := chi.NewRouter()
	logsRouter := chi.NewRouter()

	rootRouter.Use(mux.TraceIdMiddleware)

	mux.MountBoot(bootRouter)
	mux.MountImages(imgRouter)
	mux.MountKickstart(ksRouter)
	mux.MountDone(doneRouter)
	mux.MountLogs(logsRouter)
	rootRouter.Mount("/boot", bootRouter)
	rootRouter.Mount("/img", imgRouter)
	rootRouter.Mount("/ks", ksRouter)
	rootRouter.Mount("/done", doneRouter)
	rootRouter.Mount("/logs", logsRouter)
	ctl.MountServices(rootRouter)

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
			slog.ErrorContext(ctx, "shutdown error", "err", err)
		}
		close(waitForSignal)
	}()

	slog.DebugContext(ctx, "starting service", "port", config.Application.Port)

	if err := rootServer.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "listen error", "err", err)
		}
	}

	<-waitForSignal

	slog.DebugContext(ctx, "waiting for extracting jobs to complete")
	img.ExtractWG.Wait()

	slog.DebugContext(ctx, "shutdown complete")
}
