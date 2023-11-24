package img

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func Copy(ctx context.Context, destFile string, reader io.Reader) (int64, string, error) {
	file, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return 0, "", fmt.Errorf("cannot open image: %w", err)
	}
	defer file.Close()

	h := sha256.New()
	tee := io.TeeReader(reader, h)

	nBytes, err := io.Copy(file, tee)
	return nBytes, hex.EncodeToString(h.Sum(nil)), err
}
