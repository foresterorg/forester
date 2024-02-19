package main

import (
	"context"
	"forester/internal/logging"
	"forester/internal/tftp"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"golang.org/x/exp/slog"
)

const tftpTimeoutDefault = 5 * time.Second

var args struct {
	URL         string `default:"http://127.0.0.1:8000"`
	TFTPAddress string `default:":69"`
	Quiet       bool
	Verbose     bool
	Debug       bool
}

func main() {
	arg.MustParse(&args)
	if args.Debug {
		logging.Initialize(slog.LevelDebug)
	} else if args.Verbose {
		logging.Initialize(slog.LevelInfo)
	} else if args.Quiet {
		logging.Initialize(slog.LevelError)
	} else {
		logging.Initialize(slog.LevelWarn)
	}

	ctx := context.Background()
	tftp, err := tftp.Start(ctx,
		args.TFTPAddress,
		args.URL,
		tftpTimeoutDefault)
	defer tftp.Shutdown()

	if err != nil {
		slog.ErrorContext(ctx, "error when starting TFTP service", "err", err)
		os.Exit(1)
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint
	slog.DebugContext(ctx, "shutdown initiated")
}
