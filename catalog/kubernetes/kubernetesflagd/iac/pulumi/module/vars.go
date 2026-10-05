package module

var vars = struct {
	// DefaultImageRepository / DefaultImageTag are the fallbacks when the
	// platform's defaulting middleware did not run. Keep aligned with the
	// spec defaults.
	DefaultImageRepository string
	DefaultImageTag        string

	// Default ports (flagd's own defaults).
	DefaultPort           int
	DefaultManagementPort int
	DefaultSyncPort       int
	DefaultOfrepPort      int

	// MaxMetadataNameLength is the NAME BUDGET: the Service is named after
	// the resource, and a Service name is a 63-character DNS label (the
	// suffixed Secret, Role and ServiceMonitor names allow 253). Both
	// engines fail loudly beyond it (Terraform twin: a precondition).
	MaxMetadataNameLength int

	// Mount roots inside the flagd container.
	SourcesMountRoot   string
	GrpcCaMountRoot    string
	ServerTlsMountPath string
	OtelCaMountPath    string
	OtelTlsMountPath   string

	// Suffixes of module-owned object names.
	SourcesSecretSuffix  string
	FlagReaderSuffix     string
	ServiceMonitorSuffix string
}{
	DefaultImageRepository: "ghcr.io/open-feature/flagd",
	DefaultImageTag:        "v0.17.0",
	DefaultPort:            8013,
	DefaultManagementPort:  8014,
	DefaultSyncPort:        8015,
	DefaultOfrepPort:       8016,
	MaxMetadataNameLength:  63,
	SourcesMountRoot:       "/etc/flagd/sources",
	GrpcCaMountRoot:        "/etc/flagd/grpc-ca",
	ServerTlsMountPath:     "/etc/flagd/tls",
	OtelCaMountPath:        "/etc/flagd/otel-ca",
	OtelTlsMountPath:       "/etc/flagd/otel-tls",
	SourcesSecretSuffix:    "-sources",
	FlagReaderSuffix:       "-flag-reader",
	ServiceMonitorSuffix:   "-metrics",
}
