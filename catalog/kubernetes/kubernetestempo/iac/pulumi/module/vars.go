package module

// Chart identity — must stay byte-identical with the Terraform module's
// locals (helm_chart_name / helm_chart_repo): cross-engine chart-name drift
// deploys two different products from one manifest.
//
// KNOW THIS about the repo URL: the tempo chart's canonical home is the
// grafana-community index — the old https://grafana.github.io/helm-charts
// stalls at tempo 1.24.4 while the community index serves the live line.
// Never "fix" this back.
var vars = struct {
	HelmChartName string
	HelmChartRepo string
	// Fallback when spec.chart_version is unset — mirror of the proto
	// field's default option and the Terraform module's coalesce. Chart
	// 2.2.3 ships Tempo 2.10.7 (the monolithic chart); the chart pin
	// governs.
	DefaultChartVersion string
	// Fallback for spec.disk_size and spec.retention — mirrors of the
	// proto defaults.
	DefaultDiskSize  string
	DefaultRetention string
	// Deterministic env-var names carrying declared object-store
	// credentials into the Tempo config through -config.expand-env
	// (${VAR} expansion). Both engines must derive them identically.
	EnvS3AccessKeyId     string
	EnvS3SecretAccessKey string
	EnvAzureAccountKey   string
	// The r2 arm's module-owned Secret: `<name>` plus this suffix, and
	// the two keys the credential variables read (the Terraform module
	// uses the same literals).
	R2CredentialsSecretSuffix string
	R2AccessKeyIdKey          string
	R2SecretAccessKeyKey      string
	// Pod annotation carrying the fingerprint of a module-owned Secret
	// Tempo reads only at start; the same key on every catalog module
	// that uses the pattern.
	CredentialsChecksumAnnotation string
	// GCS service-account key mount contract.
	GcsKeyMountPath string
	GcsKeyVolume    string
	// Tempo's on-disk paths (the persistent volume mounts at /var/tempo).
	TracesLocalPath string
	WalPath         string
	// Service ports: the HTTP API and the two always-on OTLP receivers.
	HttpPort int
	OtlpGrpc int
	OtlpHttp int
	// The chart composes child names and truncates the fullname at 63 — a
	// resource name longer than this corrupts the naming contract the
	// outputs promise.
	MaxNameLength int
	// tempoQuery runs the docker-library-style `grafana/tempo-query`
	// image (repository-only, COMBINED form) — an image_registry override
	// re-points its repository explicitly (the tempo image is split
	// registry/repository and rides global.imageRegistry).
	TempoQueryRepository string
}{
	HelmChartName:                 "tempo",
	HelmChartRepo:                 "https://grafana-community.github.io/helm-charts",
	DefaultChartVersion:           "2.2.3",
	DefaultDiskSize:               "10Gi",
	DefaultRetention:              "24h",
	EnvS3AccessKeyId:              "TEMPO_S3_ACCESS_KEY_ID",
	EnvS3SecretAccessKey:          "TEMPO_S3_SECRET_ACCESS_KEY",
	EnvAzureAccountKey:            "TEMPO_AZURE_ACCOUNT_KEY",
	R2CredentialsSecretSuffix:     "-r2-credentials",
	R2AccessKeyIdKey:              "access-key-id",
	R2SecretAccessKeyKey:          "secret-access-key",
	CredentialsChecksumAnnotation: "checksum/credentials",
	GcsKeyMountPath:               "/var/secrets/gcs",
	GcsKeyVolume:                  "gcs-service-account",
	TracesLocalPath:               "/var/tempo/traces",
	WalPath:                       "/var/tempo/wal",
	HttpPort:                      3200,
	OtlpGrpc:                      4317,
	OtlpHttp:                      4318,
	MaxNameLength:                 45,
	TempoQueryRepository:          "grafana/tempo-query",
}
