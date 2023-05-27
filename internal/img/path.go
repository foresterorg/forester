package img

import (
	"fmt"
	"forester/internal/config"
	"os"
	"path/filepath"
	"strconv"
)

func ensureDir(imageId int64) error {
	result := filepath.Join(config.Images.Directory, strconv.FormatInt(imageId, 10))
	err := os.MkdirAll(result, 0744)
	if err != nil {
		return fmt.Errorf("cannot write image: %w", err)
	}
	return nil
}

func dirPath(imageId int64) string {
	return filepath.Join(config.Images.Directory, strconv.FormatInt(imageId, 10))
}

func isoPath(imageId int64) string {
	return filepath.Join(config.Images.Directory, strconv.FormatInt(imageId, 10), "image.iso")
}
