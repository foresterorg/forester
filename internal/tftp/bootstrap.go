package tftp

import (
	"bytes"
	"context"
	"fmt"
	"forester/internal/tmpl"
	"github.com/pin/tftp/v3"
	"io"
	"path/filepath"
	"strings"
)

func bootstrapHandler(ctx context.Context, requestPath string, rf io.ReaderFrom) error {
	if requestPath == "bootstrap/ipxe/chain.ipxe" {
		b := &bytes.Buffer{}
		err := tmpl.RenderIpxeBootstrap(ctx, b)
		if err != nil {
			return err
		}
		rf.(tftp.OutgoingTransfer).SetSize(int64(b.Len()))
		rf.ReadFrom(b)
	} else {
		filename, err := filepath.Abs(filepath.Join("/usr/share/ipxe",
			strings.TrimPrefix(requestPath, "bootstrap/ipxe/")))
		if err != nil {
			return fmt.Errorf("filepath error %s: %w", requestPath, err)
		}

		err = serveFile(filename, rf)
		if err != nil {
			return err
		}
	}
	return nil
}
	