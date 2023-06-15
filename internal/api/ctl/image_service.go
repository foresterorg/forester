package ctl

import (
	"context"
	"fmt"
	"forester/internal/config"
	"forester/internal/db"
	"forester/internal/model"
)

var _ ImageService = ImageServiceImpl{}

type ImageServiceImpl struct{}

func (i ImageServiceImpl) Create(ctx context.Context, image *Image) (int64, string, error) {
	dao := db.GetImageDao(ctx)
	dbImage := model.Image{
		Name: image.Name,
	}

	err := dao.Create(ctx, &dbImage)
	if err != nil {
		return 0, "", fmt.Errorf("cannot create: %w", err)
	}

	return dbImage.ID, fmt.Sprintf("%s/img/%d", config.BaseURL(), dbImage.ID), nil
}

func (i ImageServiceImpl) GetByID(ctx context.Context, imageID int64) (*Image, error) {
	dao := db.GetImageDao(ctx)
	result, err := dao.FindByID(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("cannot find: %w", err)
	}

	return &Image{
		ID:   result.ID,
		Name: result.Name,
	}, nil
}

func (i ImageServiceImpl) Find(ctx context.Context, pattern string) (*Image, error) {
	dao := db.GetImageDao(ctx)
	result, err := dao.Find(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("cannot find: %w", err)
	}

	return &Image{
		ID:   result.ID,
		Name: result.Name,
	}, nil
}

func (i ImageServiceImpl) List(ctx context.Context, limit int64, offset int64) ([]*Image, error) {
	dao := db.GetImageDao(ctx)
	images, err := dao.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("cannot list: %w", err)
	}
	result := make([]*Image, len(images))
	for i, img := range images {
		result[i] = &Image{
			ID:   img.ID,
			Name: img.Name,
		}
	}
	return result, nil
}

func (i ImageServiceImpl) Delete(ctx context.Context, name string) error {
	panic("implement me")
}
