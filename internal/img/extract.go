package img

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
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

type ExtractionResult struct {
	TotalSize     int64
	LiveimgSha256 string
}

func Extract(ctx context.Context, imageId int64) (*ExtractionResult, error) {
	ExtractWG.Add(1)
	defer ExtractWG.Done()
	ctx = WithJobId(ctx, NewJobId())
	result := &ExtractionResult{}

	err := ensureDir(imageId)
	if err != nil {
		slog.ErrorContext(ctx, "cannot create directory", "err", err)
		return nil, err
	}

	path := isoPath(imageId)
	slog.DebugContext(ctx, "extracting iso", "file", path)
	file, err := os.Open(path)
	if err != nil {
		slog.ErrorContext(ctx, "cannot open image", "err", err)
		return nil, err
	}
	defer file.Close()

	isoReader, err := iso9660.NewReader(file)
	if err != nil {
		slog.ErrorContext(ctx, "cannot read iso9660", "err", err)
		return nil, err
	}

	var fileCount int
	for {
		if ctx.Err() != nil {
			slog.WarnContext(ctx, "stopping extraction - context was cancelled")
			return nil, ctx.Err()
		}

		fileInfo, err := isoReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.ErrorContext(ctx, "cannot list iso9660 entry", "err", err)
			return nil, err
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
				return nil, err
			}
			continue
		}

		parentDir, _ := filepath.Split(filePath)
		if err := os.MkdirAll(parentDir, 0744); err != nil {
			slog.ErrorContext(ctx, "cannot create dir", "err", err)
			return nil, err
		}

		reader := fileInfo.Sys().(io.Reader)
		ff, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			slog.ErrorContext(ctx, "cannot open file for writing", "err", err, "file", filePath)
			return nil, err
		}

		// sha sum
		shaWriter := sha256.New()
		tee := io.TeeReader(reader, shaWriter)

		written, err := io.Copy(ff, tee)
		result.TotalSize += written

		// with or without an error from Copy, we want to attempt Close.
		closeErr := ff.Close()

		if err != nil {
			slog.ErrorContext(ctx, "iso9660 copy error", "err", err)
			return nil, err
		} else if closeErr != nil {
			slog.ErrorContext(ctx, "close error", "err", err)
			return nil, closeErr
		}

		fileCount += 1
		if fileCount > 100 {
			slog.ErrorContext(ctx, "too many files, giving up", "err", err)
			return nil, err
		}

		sha256sum := hex.EncodeToString(shaWriter.Sum(nil))
		if strings.HasSuffix(filePath, "liveimg.tar.gz") {
			result.LiveimgSha256 = sha256sum
		}

		slog.InfoContext(ctx, "extracted file", "file", path, "size", written, "sha256", sha256sum)
	}

	return result, nil
}
