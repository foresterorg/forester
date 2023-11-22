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
	ImageID     int64
	SystemID    int64
	InstallUUID string
}

type LastAction int

const (
	RebootLastAction   LastAction = iota
	PoweroffLastAction LastAction = iota
	ShutdownLastAction LastAction = iota
	HaltLastAction     LastAction = iota
)

func (la LastAction) String() string {
	switch la {
	case RebootLastAction:
		return "reboot"
	case PoweroffLastAction:
		return "poweroff"
	case ShutdownLastAction:
		return "shutdown"
	case HaltLastAction:
		return "halt"
	}
	return ""
}

type KickstartParams struct {
	*CommonParams
	ImageID        int64
	SystemID       int64
	SystemName     string
	SystemHostname string
	InstallUUID    string
	LastAction     LastAction
	Snippets       map[string][]string
	CustomSnippet  string
	LiveimgSha256  string
}

type KickstartErrorParams struct {
	*CommonParams
	Message string
}
