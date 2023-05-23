package srv

import (
	"context"
	"fmt"
	"forester/internal/db"
	"forester/internal/model"
)

func ImageList(ctx context.Context, images *[]model.Image, limit, offset int64) error {
	imageDao := db.GetImageDao(ctx)
	err := imageDao.List(ctx, images, limit, offset)
	if err != nil {
		return fmt.Errorf("image upload error: %w", err)
	}

	return nil
}
