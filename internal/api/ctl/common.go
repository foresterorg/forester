package ctl

import (
	"fmt"
	"strings"

	"forester/internal/model"
)

func ensureLimitNonzero(i *int64) {
	if i != nil && *i == 0 {
		*i = 100
	}
}

func ImageKindToInt(kind string) int16 {
	switch strings.ToLower(kind) {
	case "unknown":
		return int16(model.UnknownImageKind)
	case "image":
		return int16(model.ImageInstallerKind)
	case "container":
		return int16(model.ContainerInstallerKind)
	case "rpm":
		return int16(model.RPMInstallerKind)
	default:
		panic(fmt.Sprintf("unknown kind: %s", kind))
	}
}

func ImageIntToKind(kind int16) string {
	switch model.ImageKind(kind) {
	case model.UnknownImageKind:
		return "unknown"
	case model.ImageInstallerKind:
		return "image"
	case model.ContainerInstallerKind:
		return "container"
	case model.RPMInstallerKind:
		return "rpm"
	default:
		panic(fmt.Sprintf("unknown kind: %d", kind))
	}
}

func ApplianceKindToInt(kind string) int16 {
	switch strings.ToLower(kind) {
	case "noop":
		return int16(model.NoopApplianceKind)
	case "libvirt":
		return int16(model.LibvirtApplianceKind)
	case "redfish":
		return int16(model.RedfishApplianceKind)
	case "redfish_manual":
		return int16(model.RedfishManualApplianceKind)
	default:
		panic(fmt.Sprintf("unknown kind: %s", kind))
	}
}

func ApplianceIntToKind(kind int16) string {
	switch model.ApplianceKind(kind) {
	case model.NoopApplianceKind:
		return "noop"
	case model.LibvirtApplianceKind:
		return "libvirt"
	case model.RedfishApplianceKind:
		return "redfish"
	case model.RedfishManualApplianceKind:
		return "redfish_manual"
	default:
		panic(fmt.Sprintf("unknown kind: %d", kind))
	}
}

func SnippetKindToInt(kind string) int16 {
	switch strings.ToLower(kind) {
	case "disk":
		return 1
	case "post":
		return 2
	case "rootpw":
		return 3
	case "security":
		return 4
	case "locale":
		return 5
	case "network":
		return 6
	case "source":
		return 7
	case "debug":
		return 8
	case "pre":
		return 9
	default:
		panic(fmt.Sprintf("unknown kind: %s", kind))
	}
}

func SnippetIntToKind(kind int16) string {
	switch kind {
	case 1:
		return "disk"
	case 2:
		return "post"
	case 3:
		return "rootpw"
	case 4:
		return "security"
	case 5:
		return "locale"
	case 6:
		return "network"
	case 7:
		return "source"
	case 8:
		return "debug"
	case 9:
		return "pre"
	default:
		panic(fmt.Sprintf("unknown kind: %d", kind))
	}
}
