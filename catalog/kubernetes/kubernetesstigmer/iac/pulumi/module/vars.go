package module

import "time"

// vars holds the names this kind fixes. Every one has a twin in the Terraform
// module's locals.tf; change both together.
var vars = struct {
	// The chart, at oci://ghcr.io/stigmer/charts/stigmer-operator. Its
	// version is spec.version: the chart, the operator and the server share
	// one version line.
	ChartRepository string
	ChartName       string

	// The operator's Helm release. Fixed, as Stigmer's own guide installs it:
	// the release's cluster-wide role and the Stigmer definition it owns are
	// named after it, so a second name would be a second operator.
	OperatorRelease string

	// The Stigmer resource. Its name is fixed: every object the operator
	// creates is named after it, and the outputs read those names.
	StigmerAPIVersion string
	StigmerKind       string
	StigmerName       string
	StigmerResource   string
	StigmerDefinition string

	// The annotation that releases one held upgrade without its backup.
	UpgradeWithoutBackupAnnotation string

	// The definitions Stigmer's operator detects CloudNativePG and its backup
	// plugin by, and the label it marks what it installed itself with
	// (deploy/operator/internal/render/names.go in stigmer/stigmer).
	CloudNativePgDefinition string
	BackupPluginDefinition  string
	StigmerManagedByLabel   string
	StigmerManagedByValue   string

	// The Secret the operator keeps the encryption keys in (<name>-keys).
	KeysSecret string

	// The server's Service ports (grpc-api and artifacts).
	ServerPort    int
	ArtifactsPort int

	// The Secrets this module writes, each holding exactly the keys the spec
	// gives values for, under the names the operator reads. None collides
	// with a name the operator writes (<name>-keys, -runner-key, -postgres-*,
	// -keycloak-admin, -owner-setup, -backup-settings, -shared-runner-key).
	ModelSecret           string
	ObjectStoreSecret     string
	DatabaseSecret        string
	TemporalAPIKeySecret  string
	TemporalTLSSecret     string
	SMTPSecret            string
	LicenseSecret         string
	ServerExtraEnvSecret  string
	RunnerExtraEnvSecret  string
	DatabasePasswordKey   string
	TemporalAPIKeyKey     string
	TemporalCAKey         string
	TemporalCertKey       string
	TemporalPrivateKeyKey string
	LicenseKey            string

	// The two ConfigMaps that carry the module's cluster-side checks (gates.go).
	KeysGate   string
	WaitMarker string

	// How long a destroy waits for the operator's cleanup before it goes on
	// and says what stays. The script reports progress every half minute, so
	// a runner that stops silent jobs never sees a quiet one.
	CleanupWait time.Duration
}{
	ChartRepository:                "oci://ghcr.io/stigmer/charts",
	ChartName:                      "stigmer-operator",
	OperatorRelease:                "stigmer-operator",
	StigmerAPIVersion:              "stigmer.ai/v1",
	StigmerKind:                    "Stigmer",
	StigmerName:                    "stigmer",
	StigmerResource:                "stigmers",
	StigmerDefinition:              "stigmers.stigmer.ai",
	UpgradeWithoutBackupAnnotation: "stigmer.ai/upgrade-without-backup",
	CloudNativePgDefinition:        "clusters.postgresql.cnpg.io",
	BackupPluginDefinition:         "objectstores.barmancloud.cnpg.io",
	StigmerManagedByLabel:          "app.kubernetes.io/managed-by",
	StigmerManagedByValue:          "stigmer-operator",
	KeysSecret:                     "stigmer-keys",
	ServerPort:                     7234,
	ArtifactsPort:                  7235,
	ModelSecret:                    "stigmer-model-key",
	ObjectStoreSecret:              "stigmer-object-store-key",
	DatabaseSecret:                 "stigmer-database-password",
	TemporalAPIKeySecret:           "stigmer-temporal-api-key",
	TemporalTLSSecret:              "stigmer-temporal-tls",
	SMTPSecret:                     "stigmer-smtp-login",
	LicenseSecret:                  "stigmer-license",
	ServerExtraEnvSecret:           "stigmer-server-extra-env",
	RunnerExtraEnvSecret:           "stigmer-runner-extra-env",
	DatabasePasswordKey:            "password",
	TemporalAPIKeyKey:              "api-key",
	TemporalCAKey:                  "ca.crt",
	TemporalCertKey:                "tls.crt",
	TemporalPrivateKeyKey:          "tls.key",
	LicenseKey:                     "STIGMER_LICENSE_KEY",
	KeysGate:                       "stigmer-keys-check",
	WaitMarker:                     "stigmer-cleanup-wait",
	CleanupWait:                    10 * time.Minute,
}
