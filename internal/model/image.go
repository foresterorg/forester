package model

type Image struct {
	// Required auto-generated PK.
	ID int64 `db:"id"`

	// User-facing name. Required.
	Name string `db:"name"`

	// Image liveimg.tar.gz sha256.
	LiveimgSha256 string `db:"liveimg_sha256"`
}
