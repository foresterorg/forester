package tmpl

type CommonParams struct {
	BaseURL string
	Version string
}

type GrubErrorParams struct {
	*CommonParams
	Error error
}

type GrubKernelParams struct {
	*CommonParams
	ImageID int64
}

type KickstartParams struct {
	*CommonParams
	ImageID int64
}
