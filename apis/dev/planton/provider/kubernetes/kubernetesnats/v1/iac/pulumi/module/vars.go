package module

// vars groups together simple constants used across multiple files.
// Keeping them in a single struct mirrors Terraform's `variables.tf`.
// Note: Secret NAMES are now computed in locals.go to avoid conflicts
// when multiple instances share a namespace. Only secret KEYS are kept here.
var vars = struct {
	// NATS Helm chart info
	HelmChartName    string
	HelmChartRepoUrl string
	HelmChartVersion string

	// NACK (NATS Controllers for Kubernetes) Helm chart info
	// NACK manages JetStream resources (Streams, Consumers, etc.) via CRDs
	NackHelmChartName    string
	NackHelmChartRepoUrl string
	NackHelmChartVersion string
	// NACK app version (GitHub release tag) - differs from chart version
	NackAppVersion string
	// CRDs URL template - app version is substituted at runtime (not chart version!)
	NackCrdsUrlTemplate string

	// Fixed NATS client port
	NatsClientPort int

	// Secret keys we create / reference (note: these are keys WITHIN secrets, not secret names)
	AdminAuthSecretKey        string // bearer-token mode
	NatsUserSecretKeyUsername string // basic-auth mode
	NatsUserSecretKeyPassword string // basic-auth mode

	TlsCertKey string // key holding cert+key bundle in TLS Secret

	AdminUsername string
	// Username assigned to unauthenticated ("no-auth") clients.
	NoAuthUsername              string
	NoAuthPassword              string
	AdminUserPasswordEnvVarName string
}{
	HelmChartName:    "nats",
	HelmChartRepoUrl: "https://nats-io.github.io/k8s/helm/charts",
	HelmChartVersion: "2.12.3", // Default version, can be overridden via spec.nats_helm_chart_version

	// NACK chart and CRDs
	NackHelmChartName:    "nack",
	NackHelmChartRepoUrl: "https://nats-io.github.io/k8s/helm/charts",
	NackHelmChartVersion: "0.31.1", // Default chart version, can be overridden via spec.nack_controller.helm_chart_version
	NackAppVersion:       "0.21.1", // Default app version (GitHub tag), can be overridden via spec.nack_controller.app_version
	// CRDs are fetched using app version (not chart version!) from GitHub
	NackCrdsUrlTemplate: "https://raw.githubusercontent.com/nats-io/nack/v%s/deploy/crds.yml",

	NatsClientPort: 4222,

	// Bearer token mode: key within the secret
	AdminAuthSecretKey: "token",

	// Basic auth mode: keys within the secret
	NatsUserSecretKeyUsername: "user",
	NatsUserSecretKeyPassword: "password",

	// TLS Secret always uses the standard keys trusted by most charts:
	// { key: tls.crt / tls.key }. We export only the key name to stay within
	// a single KubernetesSecretKey field.
	TlsCertKey: "tls.crt",

	AdminUsername: "nats",

	// Fixed username for unauthenticated / anonymous clients.
	NoAuthUsername: "noauth",
	NoAuthPassword: "nopassword", // not used, but kept for consistency

	AdminUserPasswordEnvVarName: "NATS_ADMIN_PASSWORD", // env var name for admin password
}
