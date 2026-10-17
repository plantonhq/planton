package module

var vars = struct {
	// HelmChartName is the official Langfuse chart.
	HelmChartName string

	// HelmChartRepo is the served chart index.
	HelmChartRepo string

	// DefaultChartVersion is the fallback when spec.chart_version is unset
	// AND the platform's defaulting middleware did not run. Keep aligned
	// with the spec default. Each chart release pins one Langfuse version
	// as its appVersion (chart 2.1.5 = Langfuse 4.56.0).
	DefaultChartVersion string

	// MaxMetadataNameLength is the NAME BUDGET. The chart pins its fullname
	// to metadata.name (fullnameOverride) and derives every child from it;
	// the longest child is the worker PodDisruptionBudget,
	// `<fullname>-worker-pdb` (11 characters of suffix), and a Kubernetes
	// object name is a 63-character DNS label. 63 - 11 = 52, which is also
	// inside Helm's 53-character release-name limit. Both engines fail
	// loudly beyond it (Terraform twin: a lifecycle precondition on the
	// helm_release).
	MaxMetadataNameLength int

	// InstanceSecretSuffix names the module-owned Secret holding every
	// secret value the resource manages.
	InstanceSecretSuffix string

	// WebComponentSuffix names the chart's web Deployment and Service.
	WebComponentSuffix string

	// DefaultWebPort is the chart's web listener when spec.web.service.port
	// is unset.
	DefaultWebPort int

	// DefaultExternalUrl is the spec default for external_url.
	DefaultExternalUrl string

	// DefaultOrganizationName / DefaultProjectName / DefaultAdminName are the
	// spec defaults of the seeded objects.
	DefaultOrganizationName string
	DefaultProjectName      string
	DefaultAdminName        string

	// ProjectIdSuffix derives the seeded project's ID from metadata.name
	// when the spec leaves it empty.
	ProjectIdSuffix string

	// Mount directories for the files the module projects from Secrets and
	// ConfigMaps into both workloads.
	RedisTlsCaDir         string
	RedisTlsClientCertDir string
	RedisTlsClientKeyDir  string
	TrustedCaDir          string

	// InstanceSecretChecksumAnnotation carries a checksum of the declared
	// secret values on both pod templates, so a changed value (a rotated
	// SSO client secret, a new SMTP URL) rolls the pods.
	InstanceSecretChecksumAnnotation string

	// WorkerLivenessStuckPath is the worker health path that also fails
	// when queue consumption has stalled.
	WorkerLivenessStuckPath string
}{
	HelmChartName:           "langfuse",
	HelmChartRepo:           "https://langfuse.github.io/langfuse-k8s",
	DefaultChartVersion:     "2.1.5",
	MaxMetadataNameLength:   52,
	InstanceSecretSuffix:    "-secrets",
	WebComponentSuffix:      "-web",
	DefaultWebPort:          3000,
	DefaultExternalUrl:      "http://localhost:3000",
	DefaultOrganizationName: "Default Organization",
	DefaultProjectName:      "Default Project",
	DefaultAdminName:        "Admin",
	ProjectIdSuffix:         "-project",
	RedisTlsCaDir:           "/etc/langfuse/redis-tls/ca",
	RedisTlsClientCertDir:   "/etc/langfuse/redis-tls/client-cert",
	RedisTlsClientKeyDir:    "/etc/langfuse/redis-tls/client-key",
	TrustedCaDir:            "/etc/langfuse/trusted-ca",

	InstanceSecretChecksumAnnotation: "checksum/instance-secret",
	WorkerLivenessStuckPath:          "/api/health?failIfQueueConsumptionStuck=true",
}

// Keys of the instance Secret. Every key the module writes, generated or
// declared, is named here; both engines use the same names.
const (
	keySalt                       = "salt"
	keyEncryptionKey              = "encryption-key"
	keyEncryptionKeyStdB64        = "encryption-key-std-b64"
	keyNextauthSecret             = "nextauth-secret"
	keyInitProjectPublicKey       = "init-project-public-key"
	keyInitProjectSecretKey       = "init-project-secret-key"
	keyInitProjectPublicKeySuffix = "init-project-public-key-suffix"
	keyInitProjectSecretKeySuffix = "init-project-secret-key-suffix"
	keyInitUserPassword           = "init-user-password"
	keyPostgresDirectUrl          = "postgres-direct-url"
	keyPostgresShadowUrl          = "postgres-shadow-database-url"
	keyRedisTlsKeyPassphrase      = "redis-tls-key-passphrase"
	keySmtpConnectionUrl          = "smtp-connection-url"
	keyAiApiKey                   = "ai-api-key"
	keyLicenseKey                 = "license-key"
	keyAdminApiKey                = "admin-api-key"
	keySlackClientSecret          = "slack-client-secret"
	keySlackStateSecret           = "slack-state-secret"
	keyGcsCredentials             = "gcs-credentials"
	keyBlobSecretAccessKeyTmpl    = "blob-%s-secret-access-key"
	keyBlobAzureAccountKeyTmpl    = "blob-%s-azure-account-key"
	keySsoClientSecretTmpl        = "sso-%s-client-secret"
)
