package tmpl

import "forester/internal/model"

type CommonParams struct {
	BaseHost   string
	BaseURL    string
	SyslogPort string
	Version    string
}

type BootErrorType string

const (
	GrubBootErrorType BootErrorType = "grub"
	IpxeBootErrorType BootErrorType = "ipxe"
)

type BootErrorParams struct {
	*CommonParams
	Type  BootErrorType
	Error error
}

type GrubLinuxCmd string

const (
	GrubLinuxCmdBIOS   GrubLinuxCmd = "linux /boot/bios"
	GrubLinuxCmdEFIX64 GrubLinuxCmd = "linuxefi /boot/efix64"
)

type GrubInitrdCmd string

const (
	GrubInitrdCmdBIOS   GrubInitrdCmd = "initrd /boot/bios"
	GrubInitrdCmdEFIX64 GrubInitrdCmd = "initrdefi /boot/efix64"
)

type BootKernelParams struct {
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
	ImageKind      int16
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

func MakeCustomSnippets() map[string][]string {
	m := make(map[string][]string)
	for _, kind := range model.AllSnippetKinds {
		m[kind.String()] = make([]string, 0)
	}

	return m
}

type DhcpEntry struct {
	Tag string
	MAC string
}

type DhcpParams struct {
	*CommonParams
	Entries []DhcpEntry
}
