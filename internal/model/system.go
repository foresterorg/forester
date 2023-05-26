package model

import (
	"forester/internal/config"
	"time"
)

type System struct {
	// Required auto-generated PK.
	ID int64 `db:"id"`

	// Image ID or 0 when no image was assigned yet. Required.
	ImageID int64 `db:"image_id"`

	// AcquiredAt is time when system was acquired. Can be "0001-01-01 00:00:00 +0000 UTC"
	// for a system that way not acquired yet.
	AcquiredAt time.Time `db:"acquired_at"`
}

func (s System) Installable() bool {
	return time.Now().Sub(s.AcquiredAt) < config.Application.InstallDuration
}
