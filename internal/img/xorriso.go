package img

import (
	"context"
	"fmt"
	"forester/internal/logging"
	"os/exec"
	"path"
	"sync"

	"golang.org/x/exp/slog"
)

var wg sync.WaitGroup

func ExtractToDir(ctx context.Context, isoFile, outputDir string) error {
	wg.Add(1)
	defer wg.Done()

	cmd := exec.CommandContext(ctx,
		"/usr/bin/xorriso",
		"-osirrox", "on",
		"-indev", path.Clean(isoFile),
		"-extract", "/",
		path.Clean(outputDir))
	cmd.Stdout = logging.SlogWriter{Logger: slog.Default(), Level: slog.LevelDebug, Context: ctx}
	cmd.Stderr = logging.SlogWriter{Logger: slog.Default(), Level: slog.LevelWarn, Context: ctx}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error calling `xorriso` (is it installed?): %w", err)
	}

	return nil
}

func Stop() {
	wg.Wait()
}
