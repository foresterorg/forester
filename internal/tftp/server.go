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
	ctx := context.Background()
	var err error
	var i *model.Installation

	path := strings.SplitN(requestPath, "/", 2)
	mac, err := net.ParseMAC(path[0])
	if err != nil {
		return fmt.Errorf("unable to parse mac for path %s: %w", requestPath, err)
	}

	iDao := db.GetInstallationDao(ctx)
	i, _, err = iDao.FindInstallationForMAC(ctx, mac)
	if err != nil {
		return fmt.Errorf("installation not found for mac %s: %w", mac.String(), err)
	}

	if len(path) != 2 {
		return fmt.Errorf("%w: %s", ErrMalformedPath, requestPath)
	}

	root := config.BootPath(i.ImageID)
	filename, err := filepath.Abs(filepath.Join(root, path[1]))
	if err != nil {
		return fmt.Errorf("filepath error %s: %w", requestPath, err)
	}

	if !strings.HasPrefix(filename, root) {
		return ErrOutsideRoot
	}

	if strings.HasPrefix(path[1], "grub.cfg") {
		b := &bytes.Buffer{}
		mux.WriteMacConfig(ctx, b, mac, tmpl.GrubLinuxCmdBIOS, tmpl.GrubInitrdCmdBIOS)
		// set file size explicitly because buffer does not implement Seek method
		rf.(tftp.OutgoingTransfer).SetSize(int64(b.Len()))
		rf.ReadFrom(b)
	} else {
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
