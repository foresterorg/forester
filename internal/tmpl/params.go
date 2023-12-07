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

type GrubLinuxCmd string

const (
	GrubLinuxCmdBIOS GrubLinuxCmd = "linux $net_pxe_mac/"
	GrubLinuxCmdEFI GrubLinuxCmd = "linuxefi "
)

type GrubInitrdCmd string

const (
	GrubInitrdCmdBIOS GrubInitrdCmd = "initrd $net_pxe_mac/"
	GrubInitrdCmdEFI GrubInitrdCmd = "initrdefi "
)

type GrubKernelParams struct {
	*CommonParams
	ImageID     int64
	SystemID    int64
	InstallUUID string
	LinuxCmd    GrubLinuxCmd
	InitrdCmd   GrubInitrdCmd
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
