package version

var (
	// BuildCommit is SHA commit set via -ldflags
	BuildCommit string

	// BuildTime in UTC set via -ldflags
	BuildTime string
)

const (
	// ApplicationName contains string that is used by other names
	ApplicationName = "forester"

	// PrometheusLabelName contains string that is included in all Prometheus metrics
	PrometheusLabelName = ApplicationName

	// APIPathName is the name used in main route API prefix
	APIPathName = ApplicationName

	// UnleashAppName contains string used to initialize Unleash (feature flags)
	UnleashAppName = ApplicationName

	// APIPathVersion is the name used in main route API prefix
	APIPathVersion = "v1"

	// OpenTelemetryVersion is used for all OpenTelemetry tracing
	OpenTelemetryVersion = "1.0.0"
)

func init() {
	if BuildTime == "" {
		BuildTime = "N/A"
	}

	if BuildCommit == "" {
		BuildCommit = "HEAD"
	}
}
