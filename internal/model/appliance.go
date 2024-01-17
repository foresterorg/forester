package model

type Appliance struct {
	// Required auto-generated PK.
	ID int64 `db:"id"`

	// Kind is appliance type
	Kind ApplianceKind `db:"kind"`

	// User-facing name. Required.
	Name string `db:"name"`

	// URI holds connection information
	URI string `db:"uri"`
}

type ApplianceKind int16

const (
	ReservedApplianceKind      ApplianceKind = iota
	NoopApplianceKind          ApplianceKind = iota
	LibvirtApplianceKind       ApplianceKind = iota
	RedfishApplianceKind       ApplianceKind = iota
	RedfishManualApplianceKind ApplianceKind = iota
)

func ParseKind(i int16) ApplianceKind {
	switch i {
	case 0:
		return ReservedApplianceKind
	case 1:
		return NoopApplianceKind
	case 2:
		return LibvirtApplianceKind
	case 3:
		return RedfishApplianceKind
	case 4:
		return RedfishManualApplianceKind
	default:
		return -1
	}
}
