package model

type Image struct {
	// Required auto-generated PK.
	ID int64 `db:"id"`

	// User-facing name. Required.
	Name string `db:"name"`

	// Kind is image type
	Kind ImageKind `db:"kind"`

	// Image ISO SHA256.
	IsoSha256 string `db:"iso_sha256"`

	// Image liveimg.tar.gz SHA256 (when present otherwise empty string).
	LiveimgSha256 string `db:"liveimg_sha256"`
}

type ImageKind int16

const (
	UnknownImageKind  = iota
	ImageInstallerKind = iota
	ContainerInstallerKind = iota
	RPMInstallerKind = iota
)

func ParseImageKind(i int16) ApplianceKind {
	switch i {
	case 0:
		return UnknownImageKind
	case 1:
		return ImageInstallerKind
	case 2:
		return ContainerInstallerKind
	case 3:
		return RPMInstallerKind
	default:
		return -1
	}
}
