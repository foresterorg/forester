package model

type Image struct {
	// Required auto-generated PK.
	ID int64 `db:"id"`

	// User-facing name. Required.
	Name string `db:"name"`

	// Image ISO SHA256.
	IsoSha256 string `db:"iso_sha256"`

	// Image liveimg.tar.gz SHA256 (when present otherwise empty string).
	LiveimgSha256 string `db:"liveimg_sha256"`
}
