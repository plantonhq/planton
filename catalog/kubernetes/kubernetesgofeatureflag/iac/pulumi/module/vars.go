package module

var vars = struct {
	// HelmChartName is the official GO Feature Flag relay proxy chart.
	HelmChartName string

	// HelmChartRepo is the served chart index.
	HelmChartRepo string

	// DefaultChartVersion is the fallback when spec.chart_version is unset
	// AND the platform's defaulting middleware did not run. Keep aligned
	// with the spec default. Chart and relay versions move in lockstep
	// (chart 1.56.0 = relay v1.56.0).
	DefaultChartVersion string

	// MaxMetadataNameLength is the NAME BUDGET. The chart names every child
	// (Deployment, Service, ConfigMap, ServiceAccount, PDB, HPA) exactly its
	// fullname, pinned to metadata.name, and truncates the fullname at 63
	// characters - the Service name limit (a DNS label). A longer name would
	// be truncated silently and every output would name the wrong Service.
	// Both engines fail loudly instead (Terraform twin: a lifecycle
	// precondition on the helm_release).
	MaxMetadataNameLength int

	// DefaultPort / DefaultMonitoringPort are the relay listeners the spec
	// defaults to (the chart's own Service port is 1031; monitoring has no
	// chart default and is always rendered here).
	DefaultPort           int
	DefaultMonitoringPort int

	// DefaultEnvVariablePrefix scopes the environment variables the relay
	// reads as configuration. Without a prefix the relay would read EVERY
	// variable - including the service-link variables Kubernetes injects
	// for each Service in the namespace (a Service named `server` injects
	// SERVER_PORT=tcp://..., which the relay would read as server.port).
	DefaultEnvVariablePrefix string

	// FlagReaderSuffix names the Role and RoleBinding granting `get` on the
	// ConfigMaps the retrievers read; EnvSecretSuffix names the Secret
	// carrying every secret value as an environment variable;
	// ServiceMonitorSuffix names the optional ServiceMonitor.
	FlagReaderSuffix     string
	EnvSecretSuffix      string
	ServiceMonitorSuffix string

	// EnvSecretChecksumAnnotation carries the env Secret's checksum on the
	// pod template, so a changed secret value rolls the pods.
	EnvSecretChecksumAnnotation string
}{
	HelmChartName:            "relay-proxy",
	HelmChartRepo:            "https://charts.gofeatureflag.org",
	DefaultChartVersion:      "1.56.0",
	MaxMetadataNameLength:    63,
	DefaultPort:              1031,
	DefaultMonitoringPort:    1032,
	DefaultEnvVariablePrefix: "GOFFRELAY_",
	FlagReaderSuffix:         "-flag-reader",
	EnvSecretSuffix:          "-env",
	ServiceMonitorSuffix:     "-metrics",

	EnvSecretChecksumAnnotation: "checksum/env-secret",
}
