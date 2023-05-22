package img

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/config"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hooklift/iso9660"
	"golang.org/x/exp/slog"
)

var ExtractedPaths = map[string]string{
	"/_discinf.":                 "/.discinfo",
	"/efi/boot/bootx64.efi":      "/shim.efi",
	"/efi/boot/grubx64.efi":      "/grubx64.efi",
	"/liveimg_.gz":               "/liveimg.tar.gz",
	"/images/install.img":        "/images/install.img",
	"/images/pxeboot/initrd.img": "/images/pxeboot/initrd.img",
	"/images/pxeboot/vmlinuz.":   "/images/pxeboot/vmlinuz",
}

var ErrTooManyFiles = errors.New("not Anaconda image-based ISO: too many files")

func UploadImage(ctx context.Context, id int, image io.ReadSeeker) error {
	destPath := filepath.Join(config.Images.Directory, strconv.Itoa(id))
	err := os.MkdirAll(destPath, 0744)
	if err != nil {
		return fmt.Errorf("cannot write image: %w", err)
	}

	isoReader, err := iso9660.NewReader(image)
	if err != nil {
		return fmt.Errorf("cannot read iso9660: %w", err)
	}

	var fileCount int
	for {
		fileInfo, err := isoReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot list iso9660 entry: %w", err)
		}

		mappedName, ok := ExtractedPaths[fileInfo.Name()]
		if !ok {
			slog.DebugCtx(ctx, "Skipping uploaded iso9660 entry", "iso_path", fileInfo.Name())
			continue
		}

		filePath := filepath.Join(destPath, mappedName)
		slog.DebugCtx(ctx, "Extracting uploaded iso9660 entry", "iso_path", fileInfo.Name(), "dest_path", filePath)
		if fileInfo.IsDir() {
			if err := os.MkdirAll(filePath, 0744); err != nil {
				return fmt.Errorf("cannot create dir %s: %w", filePath, err)
			}
			continue
		}

		parentDir, _ := filepath.Split(filePath)
		if err := os.MkdirAll(parentDir, 0744); err != nil {
			return fmt.Errorf("cannot create dir %s: %w", parentDir, err)
		}

		reader := fileInfo.Sys().(io.Reader)
		ff, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("cannot open file %s for writing: %w", filePath, err)
		}

		_, err = io.Copy(ff, reader)

		// with or without an error from Copy, we want to attempt Close.
		cerr := ff.Close()

		if err != nil {
			return fmt.Errorf("iso9660 copy error: %w", err)
		} else if cerr != nil {
			return fmt.Errorf("file close error: %w", err)
		}

		fileCount += 1
		if fileCount > 100 {
			return ErrTooManyFiles
		}
	}

	return nil
}
