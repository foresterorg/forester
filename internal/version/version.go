package version

import "runtime/debug"

var (
	// BuildTag contains Git tag or sha
	BuildTag string

	// BuildTime contains build time
	BuildTime string

	// BuildGoVersion contains Go version
	BuildGoVersion string
)

const (
	ApplicationName = "forester"
)

func init() {
	bi, ok := debug.ReadBuildInfo()

	if ok {
		BuildGoVersion = bi.GoVersion

		for _, bs := range bi.Settings {
			switch bs.Key {
			case "vcs.revision":
				BuildTag = bs.Value[0:4]
			case "vcs.time":
				BuildTime = bs.Value
			}
		}
	}

	if BuildTag == "" {
		BuildTag = "HEAD"
	}
}
