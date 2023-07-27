package model

import (
	"forester/internal/config"
	"net"
	"strings"
	"time"
)

type System struct {
	// Required auto-generated PK.
	ID int64 `db:"id"`

	// Auto-generated human-readable name
	Name string `db:"name"`

	// Appliance ID where this system belongs or nil for manual systems
	ApplianceID *int64 `db:"appliance_id"`

	// Appliance is associated record, if present. Not all DAO functions do set this field.
	Appliance *Appliance

	// UID is unique id (typically UUID) of a system or nil for manual systems
	UID *string `db:"uid"`

	// MAC addresses
	HwAddrs HwAddrSlice `db:"hwaddrs"`

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

func (s System) Installable() bool {
	return s.Acquired && s.ImageID != nil && time.Now().Sub(s.AcquiredAt) < config.Application.InstallDuration
}

func (s System) UniqueHwAddrs() []net.HardwareAddr {
	return s.HwAddrs.Unique()
}

func (s System) HwAddrStrings() []string {
	hwa := s.UniqueHwAddrs()
	result := make([]string, len(hwa))
	for i, a := range hwa {
		result[i] = a.String()
	}
	return result
}

func (s System) HwAddrString() string {
	return strings.Join(s.HwAddrStrings(), ",")
}

type SystemAppliance struct {
	System    `db:"s"`
	Appliance `db:"a"`
}
