package img

import (
	"context"
	"fmt"
	"forester/internal/logging"
	"os"
	"os/exec"
	"path"

	"golang.org/x/exp/slog"
)

func GenerateBootISO(ctx context.Context, imageID int64, imageDir string) error {
	wg.Add(1)
	defer wg.Done()

	cmd := exec.CommandContext(ctx, "/usr/bin/bash")
	cmd.Stdout = logging.SlogWriter{Logger: slog.Default(), Level: slog.LevelDebug, Context: ctx}
	cmd.Stderr = logging.SlogWriter{Logger: slog.Default(), Level: slog.LevelWarn, Context: ctx}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error opening stdin of shell: %w", err)
	}
	err = renderGenerateBootISO(ctx, stdin, imageID, imageDir)
	if err != nil {
		return fmt.Errorf("error rendering boot.iso generator script template: %w", err)
	}
	stdin.Close()

	err = cmd.Run()
	if err != nil {
		script := path.Join(imageDir, "genboot.sh")
		f, ferr := os.OpenFile(script, os.O_CREATE|os.O_WRONLY, 0770)
		if ferr == nil {
			defer f.Close()
			slog.WarnContext(ctx, "writing failed generate ISO script", "file", script, "err", err)
			renderGenerateBootISO(ctx, f, imageID, imageDir)
		}
		return fmt.Errorf("error calling ISO generator script: %w", err)
	}

	return nil
}
