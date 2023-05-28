package model

import (
	"forester/internal/config"
	"net"
	"time"
)

type System struct {
	// Required auto-generated PK.
	ID int64 `db:"id"`

	// Auto-generated human-readable name
	Name string `db:"name"`

	// MAC addresses
	HwAddrs []net.HardwareAddr `db:"hwaddrs"`

	// Details about the system
	Facts Facts `db:"facts"`

	// Whether a system is owned by someone
	Acquired bool `db:"acquired"`

	// AcquiredAt is time when system was acquired. Can be "0001-01-01 00:00:00 +0000 UTC"
	// for a system that way not acquired yet.
	AcquiredAt time.Time `db:"acquired_at"`

	// Image ID or 0 when no image was assigned yet.
	ImageID int64 `db:"image_id"`

	// Comment, can be blank.
	Comment string `db:"comment"`
}

func (s System) Installable() bool {
	return time.Now().Sub(s.AcquiredAt) < config.Application.InstallDuration
}
