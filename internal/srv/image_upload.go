package srv

import (
	"context"
	"errors"
	"fmt"
	"forester/internal/db"
	"forester/internal/img"
	"forester/internal/model"
	"io"
)

var ErrImageNameMissing = errors.New("missing image name")

func ImageUpload(ctx context.Context, contents io.ReadSeeker, imageName string) error {
	if imageName == "" {
		return ErrImageNameMissing
	}
	imageDao := db.GetImageDao(ctx)
	imgModel := model.Image{
		Name: imageName,
	}
	err := imageDao.Create(ctx, &imgModel)
	if err != nil {
		return fmt.Errorf("image upload error: %w", err)
	}
	err = img.UploadImage(ctx, imgModel.ID, contents)
	if err != nil {
		return fmt.Errorf("image upload error: %w", err)
	}

	return nil
}
