package img

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/hooklift/iso9660"
	"golang.org/x/exp/slog"
)

// ExtractedPaths is a mapping between files and target names. Some images do not
// use iso9660 extensions and therefore longer names may be shortened.
var ExtractedPaths = map[string]string{
	"/_discinfo":                 "/.discinfo",
	"/_discinf.":                 "/.discinfo",
	"/efi/boot/bootx64.efi":      "/shim.efi",
	"/efi/boot/grubx64.efi":      "/grubx64.efi",
	"/liveimg.tar.gz":            "/liveimg.tar.gz",
	"/liveimg_tar.gz":            "/liveimg.tar.gz",
	"/liveimg_.gz":               "/liveimg.tar.gz",
	"/images/install.img":        "/images/install.img",
	"/images/pxeboot/initrd.img": "/images/pxeboot/initrd.img",
	"/images/pxeboot/vmlinuz":    "/images/pxeboot/vmlinuz",
	"/images/pxeboot/vmlinuz.":   "/images/pxeboot/vmlinuz",
}

var ExtractWG sync.WaitGroup

func Extract(ctx context.Context, imageId int64) {
	ExtractWG.Add(1)
	defer ExtractWG.Done()
	ctx = WithJobId(ctx, NewJobId())

	err := ensureDir(imageId)
	if err != nil {
		slog.ErrorContext(ctx, "cannot create directory", "err", err)
		return
	}

	path := isoPath(imageId)
	slog.DebugContext(ctx, "extracting iso", "file", path)
	file, err := os.Open(path)
	if err != nil {
		slog.ErrorContext(ctx, "cannot open image", "err", err)
		return
	}
	defer file.Close()

	isoReader, err := iso9660.NewReader(file)
	if err != nil {
		slog.ErrorContext(ctx, "cannot read iso9660", "err", err)
		return
	}

	var fileCount int
	for {
		fileInfo, err := isoReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.ErrorContext(ctx, "cannot list iso9660 entry", "err", err)
			return
		}

		mappedName, ok := ExtractedPaths[fileInfo.Name()]
		if !ok {
			slog.DebugContext(ctx, "skipping uploaded iso9660 entry", "iso_path", fileInfo.Name())
			continue
		}

		filePath := filepath.Join(dirPath(imageId), mappedName)
		slog.DebugContext(ctx, "extracting uploaded iso9660 entry", "iso_path", fileInfo.Name(), "dest_path", filePath)
		if fileInfo.IsDir() {
			if err := os.MkdirAll(filePath, 0744); err != nil {
				slog.ErrorContext(ctx, "cannot create dir", "err", err)
				return
			}
			continue
		}

		parentDir, _ := filepath.Split(filePath)
		if err := os.MkdirAll(parentDir, 0744); err != nil {
			slog.ErrorContext(ctx, "cannot create dir", "err", err)
			return
		}

		reader := fileInfo.Sys().(io.Reader)
		ff, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			slog.ErrorContext(ctx, "cannot open file for writing", "err", err, "file", filePath)
			return
		}

		_, err = io.Copy(ff, reader)

		// with or without an error from Copy, we want to attempt Close.
		closeErr := ff.Close()

		if err != nil {
			slog.ErrorContext(ctx, "iso9660 copy error", "err", err)
			return
		} else if closeErr != nil {
			slog.ErrorContext(ctx, "close error", "err", err)
			return
		}

		fileCount += 1
		if fileCount > 100 {
			slog.ErrorContext(ctx, "too many files, giving up", "err", err)
			return
		}

		slog.InfoContext(ctx, "extraction done", "file", path)
	}
}
