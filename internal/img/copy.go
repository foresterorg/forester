package img

import (
	"context"
	"fmt"
	"io"
	"os"
)

func Copy(ctx context.Context, imageId int64, reader io.Reader) (int64, error) {
	err := ensureDir(imageId)
	if err != nil {
		return 0, fmt.Errorf("cannot create directory: %w", err)
	}

	file, err := os.OpenFile(isoPath(imageId), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return 0, fmt.Errorf("cannot open image: %w", err)
	}
	defer file.Close()

	nBytes, err := io.Copy(file, reader)
	return nBytes, err
}
