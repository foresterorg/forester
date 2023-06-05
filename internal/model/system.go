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

	// Image ID or nil when no image was acquired yet.
	ImageID *int64 `db:"image_id"`

	// Comment, can be blank.
	Comment string `db:"comment"`
}

type Fact struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Facts struct {
	List []Fact `json:"list"`
}

func (f *Facts) FactsMap() map[string]string {
	result := make(map[string]string, len(f.List))
	for _, f := range f.List {
		result[f.Key] = f.Value
	}
	return result
}

var KnownFactKeys = [...]string{
	"bios_vendor",
	"bios_version",
	"bios_release_date",
	"bios_revision",
	"firmware_revision",
	"system_manufacturer",
	"system_product_name",
	"system_version",
	"system_serial_number",
	"system_uuid",
	"system_sku_number",
	"system_family",
	"baseboard_manufacturer",
	"baseboard_product_name",
	"baseboard_version",
	"baseboard_serial_number",
	"baseboard_asset_tag",
	"chassis_manufacturer",
	"chassis_type",
	"chassis_version",
	"chassis_serial_number",
	"chassis_asset_tag",
	"processor_family",
	"processor_manufacturer",
	"processor_version",
	"processor_frequency",
}

func (s System) Installable() bool {
	return s.Acquired && s.ImageID != nil && time.Now().Sub(s.AcquiredAt) < config.Application.InstallDuration
}

func (s System) HwAddrStrings() []string {
	result := make([]string, len(s.HwAddrs))
	for i, a := range s.HwAddrs {
		result[i] = a.String()
	}
	return result
}
