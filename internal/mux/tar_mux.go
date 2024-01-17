package mux

import (
	"archive/tar"
	"forester/internal/config"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slog"
)

func MountTar(r *chi.Mux) {
	r.Head("/{ID}/container", serveContainerTar)
	r.Get("/{ID}/container", serveContainerTar)
}

func serveContainerTar(w http.ResponseWriter, r *http.Request) {
	imgID, err := strconv.ParseInt(chi.URLParam(r, "ID"), 10, 64)
	if err != nil {
		slog.WarnContext(r.Context(), "cannot parse ID", "err", err)
		http.NotFound(w, r)
		return
	}

	root := path.Clean(path.Join(config.BootPath(imgID), "container"))
	stat, err := os.Stat(root)
	if err != nil {
		slog.WarnContext(r.Context(), "directory for image does not exist", "path", root, "err", err)
		http.NotFound(w, r)
		return
	}

	if !stat.IsDir() {
		slog.WarnContext(r.Context(), "image is not container type", "path", root, "ID", imgID, "err", err)
		http.NotFound(w, r)
		return
	}

	slog.DebugContext(r.Context(), "serving tar", "path", root)
	w.Header().Add("Content-Type", "application/x-tar")
	tw := tar.NewWriter(w)
	defer tw.Close()

	filepath.WalkDir(root, func(path string, de fs.DirEntry, err error) error {
		if err != nil {
			slog.WarnContext(r.Context(), "walk error", "path", path, "err", err)
			return err
		}

		if de.IsDir() {
			return nil
		}

		tarName := strings.TrimPrefix(path, root+"/")
		//slog.DebugContext(r.Context(), tarName, "path", root)

		fi, err := de.Info()
		if err != nil {
			slog.WarnContext(r.Context(), "cannot get file info", "path", path, "err", err)
			return err
		}

		header := &tar.Header{
			Name:    tarName,
			Size:    fi.Size(),
			Mode:    int64(fi.Mode()),
			ModTime: fi.ModTime(),
		}

		err = tw.WriteHeader(header)
		if err != nil {
			slog.WarnContext(r.Context(), "cannot write tar header", "path", path, "err", err)
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			slog.WarnContext(r.Context(), "cannot open file", "path", path, "err", err)
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		if err != nil {
			slog.WarnContext(r.Context(), "cannot write file contents", "path", path, "err", err)
			return err
		}

		return nil
	})

}
