package tftp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pin/tftp/v3"
	"golang.org/x/exp/slog"
)

type Server struct {
	ts  *tftp.Server
	url string
	c   *http.Client
}

var ErrOutsideRoot = errors.New("access outside of the root directory")
var ErrMalformedPath = errors.New("malformed path")

func urlJoin(base string, other string) (string, error) {
	if !strings.HasSuffix(base, "/") {
		base = base + "/"
	}

	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	o, err := url.Parse(strings.TrimPrefix(other, "/"))
	if err != nil {
		return "", err
	}

	u := b.ResolveReference(o)
	return u.String(), nil
}

var ErrNotSupported = errors.New("not supported")
var ErrNotFound = errors.New("file not found")
var ErrUnknown = errors.New("unknown error")

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

func (s *Server) readHandler() func(filename string, rf io.ReaderFrom) error {
	return func(filename string, rf io.ReaderFrom) error {
		uri, err := urlJoin(s.url, filename)
		if err != nil {
			return fmt.Errorf("error building URL: %w", err)
		}

		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			return fmt.Errorf("cannot create HTTP request: %w", err)
		}

		raddr := rf.(tftp.OutgoingTransfer).RemoteAddr()
		req.Header.Add("X-Tftp-Ip", raddr.IP.String())
		req.Header.Add("X-Tftp-Port", fmt.Sprintf("%d", raddr.Port))
		req.Header.Add("X-Tftp-File", filename)

		resp, err := s.c.Do(req)
		if err != nil {
			return fmt.Errorf("error during HTTP request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return ErrNotFound
		} else if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%w: %s", ErrUnknown, resp.Status)
		}

		// Use ContentLength, if provided, to set TSize option
		if resp.ContentLength >= 0 {
			rf.(tftp.OutgoingTransfer).SetSize(resp.ContentLength)
		}

		_, err = rf.ReadFrom(resp.Body)
		if err != nil {
			return fmt.Errorf("readfrom failed: %w", err)
		}

		return nil
	}
}

func Start(ctx context.Context, listenAddress, url string, timeout time.Duration) (*Server, error) {
	server := &Server{
		c: &http.Client{},
		url: url,
	}
	slog.InfoContext(ctx, "starting TFTP server",
		"address", listenAddress,
		"url", url,
		"timeout", timeout)

	server.ts = tftp.NewServer(server.readHandler(), writeHandler)
	server.ts.SetHook(&logHook{})
	server.ts.SetTimeout(timeout)

	go func() {
		err := server.ts.ListenAndServe(listenAddress)
		if err != nil {
			slog.ErrorContext(ctx, "error when starting TFTP service", "address", listenAddress, "err", err)
			os.Exit(1)
		}
	}()

	return server, nil
}

func (s *Server) Shutdown() {
	s.ts.Shutdown()
}
