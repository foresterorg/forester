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
	"net/netip"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
	"gopkg.in/mcuadros/go-syslog.v2"
)

func startSyslog(ctx context.Context, server *syslog.Server) {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

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

	err = os.MkdirAll(config.Logging.SyslogDir, 0x755)
	if err != nil {
		fmt.Printf("Cannot create syslog dir %s: %s", config.Logging.SyslogDir, err.Error())
		os.Exit(1)
	}

	go syslogHandler(ctx, channel)
}

func closeFiles(files map[netip.Addr]*os.File) {
	for k, f := range files {
		if f == nil {
			delete(files, k)
			continue
		}
		slog.Debug("closing syslog file", "file", f.Name())
		err := f.Close()
		if err != nil {
			slog.Error("cannot close", "file", f.Name(), "err", err.Error())
		}
		delete(files, k)
	}
}

func syslogHandler(ctx context.Context, channel syslog.LogPartsChannel) {
	files := make(map[netip.Addr]*os.File)
	defer closeFiles(files)
	closeTicker := time.Tick(time.Second * 5)

	for {
		select {
		case logParts := <-channel:
			client, ok := logParts["client"]
			if !ok {
				client = "0.0.0.0:0"
			}
			ap, err := netip.ParseAddrPort(client.(string))
			if err != nil {
				slog.ErrorContext(ctx, "cannot parse syslog client field", "err", err.Error())
				continue
			}
			f, ok := files[ap.Addr()]
			if !ok {
				var err error
				name := fmt.Sprintf("%s_%s.log", time.Now().Format("2006-02-01"), ap.Addr().String())
				slog.DebugContext(ctx, "opening syslog file", "file", name)
				fp := path.Join(config.Logging.SyslogDir, name)
				f, err = os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					slog.ErrorContext(ctx, "cannot open file for appending", "file", fp, "err", err.Error())
				}
				files[ap.Addr()] = f
				slog.DebugContext(ctx, "file map", "size", len(files))
			}

			if _, err := f.WriteString(fmt.Sprintf("%s\n", logParts["content"])); err != nil {
				slog.ErrorContext(ctx, "cannot append to file", "file", f.Name(), "err", err.Error())
			}

			if config.Logging.Syslog {
				var attrs []slog.Attr
				for k, v := range logParts {
					if k != "content" {
						attrs = append(attrs, slog.Any(k, v))
					}
				}
				slog.DebugContext(ctx, fmt.Sprintf("%s", logParts["content"]), "syslog", attrs)
			}
		case <-closeTicker:
			closeFiles(files)
		case <-ctx.Done():
			return
		}
	}
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

	syslogServer := syslog.NewServer()
	if config.Logging.Syslog {
		syslogCtx, syslogCancel := context.WithCancel(ctx)
		defer syslogCancel()
		startSyslog(syslogCtx, syslogServer)
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

	rootRouter.Use(mux.TraceIdMiddleware)

	mux.MountBoot(bootRouter)
	mux.MountImages(imgRouter)
	mux.MountKickstart(ksRouter)
	mux.MountDone(doneRouter)
	rootRouter.Mount("/boot", bootRouter)
	rootRouter.Mount("/img", imgRouter)
	rootRouter.Mount("/ks", ksRouter)
	rootRouter.Mount("/done", doneRouter)
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

	slog.DebugContext(ctx, "stopping syslog listeners")
	err = syslogServer.Kill()
	if err != nil {
		slog.ErrorContext(ctx, "cannot stop syslog server", "err", err)
	}
	syslogServer.Wait()

	slog.DebugContext(ctx, "shutdown complete")
}
