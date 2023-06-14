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
	ReservedKind = iota
	LibvirtKind  = iota
)

func ParseKind(i int16) ApplianceKind {
	switch i {
	case 0:
		return ReservedKind
	case 1:
		return LibvirtKind
	default:
		return -1
	}
}
