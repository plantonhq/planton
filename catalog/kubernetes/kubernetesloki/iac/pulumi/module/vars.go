package module

// Chart identity — must stay byte-identical with the Terraform module's
// locals (helm_chart_name / helm_chart_repo): cross-engine chart-name
// drift deploys two different products from one manifest.
//
// KNOW THIS about the repo URL: the loki chart's canonical home is the
// grafana-community index (the vendor moved its community charts there —
// the same move the grafana chart made). Never "fix" this back to the
// grafana.github.io URL.
var vars = struct {
	HelmChartName string
	HelmChartRepo string
	// Fallback when spec.chart_version is unset — mirror of the proto
	// field's default option and the Terraform module's coalesce. Chart
	// 18.5.4 ships Loki 3.7.4; the chart pin governs.
	DefaultChartVersion string
	// Fallback for the modes' disk_size — mirror of the proto defaults.
	DefaultDiskSize string
	// The derived schema's start date for NEW installs. Any fixed past
	// date works (the schema only has to start before the first ingested
	// log); this is the chart's own reference-values date. Imports of
	// existing clusters override it via spec.schema_from_date.
	DefaultSchemaFromDate string
	// The one index-schema generation this component renders: TSDB on
	// schema v13 with 24h index periods — the current-generation Loki
	// schema. Upstream makes every user hand-author this block; deriving
	// it is this kind's ergonomic win.
	SchemaStore       string
	SchemaVersion     string
	SchemaIndexPrefix string
	SchemaIndexPeriod string
	// Deterministic env-var names carrying declared object-store
	// credentials into the Loki config through -config.expand-env
	// (${VAR} expansion). The names appear in the rendered config AND
	// the pods' env — both engines must derive them identically.
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
	// Loki reads only at start; the same key on every catalog module
	// that uses the pattern.
	CredentialsChecksumAnnotation string
	// GCS service-account key mount contract (the key file is mounted
	// from the referenced Secret and named through
	// GOOGLE_APPLICATION_CREDENTIALS).
	GcsKeyMountPath string
	GcsKeyVolume    string
	// The gateway's Service port (nginx listens on 8080 behind a port-80
	// Service) and Loki's own HTTP port.
	GatewayServicePort int
	LokiHttpPort       int
	// The chart's name budget: it composes child names like
	// `<fullname>-backend-headless` (16 chars of suffix) and truncates
	// the fullname at 63 — a resource name longer than this corrupts the
	// naming contract the outputs promise.
	MaxNameLength int
	// The memcached caches run the docker-library `memcached` image
	// (repository-only, COMBINED form) while every other chart image is
	// split registry/repository — an image_registry override must
	// handle both forms (see values.go).
	MemcachedRepository string
}{
	HelmChartName:                 "loki",
	HelmChartRepo:                 "https://grafana-community.github.io/helm-charts",
	DefaultChartVersion:           "18.5.4",
	DefaultDiskSize:               "10Gi",
	DefaultSchemaFromDate:         "2024-04-01",
	SchemaStore:                   "tsdb",
	SchemaVersion:                 "v13",
	SchemaIndexPrefix:             "loki_index_",
	SchemaIndexPeriod:             "24h",
	EnvS3AccessKeyId:              "LOKI_S3_ACCESS_KEY_ID",
	EnvS3SecretAccessKey:          "LOKI_S3_SECRET_ACCESS_KEY",
	EnvAzureAccountKey:            "LOKI_AZURE_ACCOUNT_KEY",
	R2CredentialsSecretSuffix:     "-r2-credentials",
	R2AccessKeyIdKey:              "access-key-id",
	R2SecretAccessKeyKey:          "secret-access-key",
	CredentialsChecksumAnnotation: "checksum/credentials",
	GcsKeyMountPath:               "/var/secrets/gcs",
	GcsKeyVolume:                  "gcs-service-account",
	GatewayServicePort:            80,
	LokiHttpPort:                  3100,
	MaxNameLength:                 40,
	MemcachedRepository:           "memcached",
}

// distributedComponents are the chart's microservices-mode workloads,
// zeroed in EVERY rendering: the chart's own reference values files zero
// them by hand for the monolithic and simple-scalable modes, and leaving
// the zeroing to the operator's care is exactly the half-running-mode trap
// the spec promises to prevent.
var distributedComponents = []string{
	"ingester",
	"querier",
	"queryFrontend",
	"queryScheduler",
	"distributor",
	"compactor",
	"indexGateway",
	"bloomCompactor",
	"bloomGateway",
}
