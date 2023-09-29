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
	"forester/internal/mux"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"gopkg.in/mcuadros/go-syslog.v2"
)

func syslogd() {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)
	err := server.ListenUDP(fmt.Sprintf("0.0.0.0:%d", config.Application.SyslogPort))
	if err != nil {
		fmt.Printf("Cannot listen on UDP port %d: %s", config.Application.SyslogPort, err.Error())
		os.Exit(1)
	}
	err = server.ListenTCP(fmt.Sprintf("0.0.0.0:%d", config.Application.SyslogPort))
	if err != nil {
		fmt.Printf("Cannot listen on TCP port %d: %s", config.Application.SyslogPort, err.Error())
		os.Exit(1)
	}
	err = server.Boot()
	if err != nil {
		fmt.Printf("Cannot start syslog server on UDP port %d: %s", config.Application.SyslogPort, err.Error())
		os.Exit(1)
	}

	go func(channel syslog.LogPartsChannel) {
		// In the future, some ring-buffer could allow displaying or following live logs. Alternatively,
		// logs could be stored in text files under hostname/session.log format where session would be
		// hash of syslog.client field (IP:PORT). Send logs to the app logger for now.
		for logParts := range channel {
			var attrs []slog.Attr
			for k, v := range logParts {
				if k != "content" {
					attrs = append(attrs, slog.Any(k, v))
				}
			}
			slog.Debug(fmt.Sprintf("%s", logParts["content"]), "syslog", attrs)
		}
	}(channel)

	server.Wait()
}

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

	go syslogd()

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

	rootRouter.Use(mux.TraceIdMiddleware)

	mux.MountBoot(bootRouter)
	mux.MountImages(imgRouter)
	mux.MountKickstart(ksRouter)
	rootRouter.Mount("/boot", bootRouter)
	rootRouter.Mount("/img", imgRouter)
	rootRouter.Mount("/ks", ksRouter)
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
