package db

import (
	"context"
	"forester/internal/model"
	"time"
)

func init() {
	GetSystemDao = getSystemDao
}

type systemDao struct{}

func getSystemDao(ctx context.Context) SystemDao {
	return &systemDao{}
}

func (dao systemDao) FindByMac(ctx context.Context, mac string) (*model.System, error) {
	result := &model.System{}
	result.ImageID = 5
	result.AcquiredAt = time.Now()
	return result, nil
}
