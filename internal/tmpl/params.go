package tmpl

type CommonParams struct {
	BaseHost   string
	BaseURL    string
	SyslogPort string
	Version    string
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

type KickstartErrorParams struct {
	*CommonParams
	Message string
}
