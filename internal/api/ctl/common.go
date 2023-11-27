package ctl

import (
	"fmt"
	"strings"
)

func ensureLimitNonzero(i *int64) {
	if i != nil && *i == 0 {
		*i = 100
	}
}

func ApplianceKindToInt(kind string) int16 {
	switch strings.ToLower(kind) {
	case "noop":
		return 1
	case "libvirt":
		return 2
	case "redfish":
		return 3
	case "redfish_manual":
		return 4
	default:
		panic(fmt.Sprintf("unknown kind: %s", kind))
	}
}

func ApplianceIntToKind(kind int16) string {
	switch kind {
	case 1:
		return "noop"
	case 2:
		return "libvirt"
	case 3:
		return "redfish"
	case 4:
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
