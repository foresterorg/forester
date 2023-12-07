package tftp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"forester/internal/config"
	"forester/internal/db"
	"forester/internal/model"
	"forester/internal/mux"
	"forester/internal/tmpl"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pin/tftp/v3"
	"golang.org/x/exp/slog"
)

type Server struct {
	ts *tftp.Server
}

var ErrOutsideRoot = errors.New("access outside of the root directory")
var ErrMalformedPath = errors.New("malformed path")

func readHandler(requestPath string, rf io.ReaderFrom) error {
	slog.Debug("serving", "path", requestPath)
	requestPath = strings.TrimPrefix(requestPath, "/")
	ctx := context.Background()
	var err error
	var i *model.Installation

	path := strings.SplitN(requestPath, "/", 3)
	if len(path) != 3 || path[0] == "" {
		return fmt.Errorf("%w: %+v", ErrMalformedPath, path)
	}

	mac, err := net.ParseMAC(path[1])
	if err != nil {
		return fmt.Errorf("unable to parse mac %s for path %s: %w", path[1], requestPath, err)
	}

	platform := strings.ToLower(path[0])
	finalPath := path[2]

	iDao := db.GetInstallationDao(ctx)
	i, _, err = iDao.FindInstallationForMAC(ctx, mac)
	if err != nil {
		return fmt.Errorf("installation not found for mac %s: %w", mac.String(), err)
	}

	root := config.BootPath(i.ImageID)
	filename, err := filepath.Abs(filepath.Join(root, finalPath))
	if err != nil {
		return fmt.Errorf("filepath error %s: %w", requestPath, err)
	}

	if !strings.HasPrefix(filename, root) {
		return ErrOutsideRoot
	}

	if strings.HasPrefix(finalPath, "grub.cfg") {
		b := &bytes.Buffer{}
		if platform == "bios" {
			mux.WriteMacConfig(ctx, b, mac, tmpl.GrubLinuxCmdBIOS, tmpl.GrubInitrdCmdBIOS)
		} else {
			mux.WriteMacConfig(ctx, b, mac, tmpl.GrubLinuxCmdEFIX64, tmpl.GrubInitrdCmdEFIX64)
		}
		// set file size explicitly because buffer does not implement Seek method
		rf.(tftp.OutgoingTransfer).SetSize(int64(b.Len()))
		rf.ReadFrom(b)
	} else {
		slog.Debug("serving file", "file", filename, "mac", mac.String(), "platform", platform)
		file, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("cannot open %s: %w", requestPath, err)
		}
		defer file.Close()
		_, err = rf.ReadFrom(file)
		if err != nil {
			return fmt.Errorf("cannot read from %s: %w", requestPath, err)
		}
	}

	return nil
}

var ErrNotSupported = errors.New("writing not supported")

func writeHandler(filename string, wt io.WriterTo) error {
	return ErrNotSupported
}

// Hook for logging on every transfer completion or failure.
type logHook struct{}

func (h *logHook) OnSuccess(s tftp.TransferStats) {
	slog.Info("tftp transfer complete",
		"file", s.Filename,
		"remote", s.RemoteAddr,
		"duration", s.Duration,
		"dack", s.DatagramsAcked,
		"dsnt", s.DatagramsSent,
	)
}

func (h *logHook) OnFailure(s tftp.TransferStats, err error) {
	slog.Info("tftp transfer complete",
		"err", err,
		"file", s.Filename,
		"remote", s.RemoteAddr,
		"duration", s.Duration,
		"dack", s.DatagramsAcked,
		"dsnt", s.DatagramsSent,
	)
}

func Start(ctx context.Context) (*Server, error) {
	server := &Server{}
	slog.InfoContext(ctx, "starting TFTP server", "port", config.Tftp.Port)

	server.ts = tftp.NewServer(readHandler, writeHandler)
	server.ts.SetHook(&logHook{})

	go func() {
		err := server.ts.ListenAndServe(fmt.Sprintf(":%d", config.Tftp.Port))
		if err != nil {
			slog.ErrorContext(ctx, "error when starting TFTP service", "err", err)
			os.Exit(1)
		}
	}()

	return server, nil
}

func (s *Server) Shutdown() {
	s.ts.Shutdown()
}
