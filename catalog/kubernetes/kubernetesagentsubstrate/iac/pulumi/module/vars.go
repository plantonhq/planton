package module

// vars holds the upstream contract the module installs: the release whose
// install it composes, where releases are served, and the fixed names
// upstream's manifests bake in. Every value has an exact twin in the
// Terraform module's locals.tf, and a test holds the two equal.
var vars = struct {
	// Release is the agent-substrate/substrate release installed when
	// spec.version is empty, and the one the composition and its fixtures
	// (testdata/<Release>) are written against. Any other release installs
	// with this module as long as upstream keeps what the composition reads
	// (release.go); moving the default is an optional catalog update, and the
	// upgrade rule (upgrade.go) says which installed releases it moves from.
	Release string

	// BundleName is the upstream project's name, the ownership label the
	// kept-CRD primitive stamps on the definitions.
	BundleName string

	// Namespace holds the control plane; PodCertificatesNamespace holds the
	// pod-certificate controller and its CA pools. Both are baked into
	// upstream's manifests.
	Namespace                string
	PodCertificatesNamespace string

	// The workloads the module applies under upstream's fixed names, each
	// ownership-checked before anything is written.
	APIServerName          string
	ControllerName         string
	NodeAgentName          string
	RouterName             string
	EgressName             string
	PodCertificatesName    string
	CredentialProviderName string
	PostgresName           string
	IdentityServerName     string

	// PostgresVolumeClaim is the bundled database's volume, created by the
	// StatefulSet from its "data" claim template; it outlives a destroy while
	// another resource owns the namespace.
	PostgresVolumeClaim string

	// The API surface a product reads (outputs).
	APIServerHost       string
	APITrustBundleName  string
	APITrustBundleSign  string
	APIAccessLabelKey   string
	APIAccessLabelValue string
	// RouterClientSigner issues the client certificates the router's TLS
	// endpoint accepts (upstream's podidentity signer).
	RouterClientSigner string

	// VersionAnnotation records the installed release on the API server
	// Deployment; MinorLabel records the minor that wrote the bundled
	// database on its claim template, so the claim carries it past a destroy.
	// Upstream writes no version marker, and the upgrade refusal reads these.
	VersionAnnotation string
	MinorLabel        string

	// EnvHashAnnotation is upstream's own: the digest of the API server's
	// environment ConfigMap and Secret on its pod template, so a change rolls
	// it. ConfigHashAnnotation is the module's: the digest of the other
	// spec-derived configuration a workload reads only at start.
	EnvHashAnnotation    string
	ConfigHashAnnotation string

	// PostgresSetupConfigMap holds upstream's setup script for the bundled
	// database, mounted at PostgresSetupMountPath and run by the database's
	// own start hook (the installer runs it by exec on every deploy, which
	// no IaC engine can).
	PostgresSetupConfigMap string
	PostgresSetupMountPath string
}{
	Release:                  "v1.0.0",
	BundleName:               "agent-substrate",
	Namespace:                "ate-system",
	PodCertificatesNamespace: "podcertificate-controller-system",
	APIServerName:            "ate-api-server",
	ControllerName:           "ate-controller",
	NodeAgentName:            "atelet",
	RouterName:               "atenet-router",
	EgressName:               "atenet-egress",
	PodCertificatesName:      "podcertificate-controller",
	CredentialProviderName:   "k8s-credential-provider",
	PostgresName:             "postgres",
	IdentityServerName:       "ate-idp-server",
	PostgresVolumeClaim:      "data-postgres-0",
	APIServerHost:            "api.ate-system.svc",
	APITrustBundleName:       "servicedns.podcert.ate.dev:identity:primary-bundle",
	APITrustBundleSign:       "servicedns.podcert.ate.dev/identity",
	APIAccessLabelKey:        "ate.dev/api-access",
	APIAccessLabelValue:      "true",
	RouterClientSigner:       "podidentity.podcert.ate.dev/identity",
	VersionAnnotation:        "planton.ai/agent-substrate-version",
	MinorLabel:               "planton.ai/agent-substrate-minor",
	EnvHashAnnotation:        "ate.dev/env-hash",
	ConfigHashAnnotation:     "planton.ai/config-hash",
	PostgresSetupConfigMap:   "postgres-setup",
	PostgresSetupMountPath:   "/etc/postgres-setup",
}

// The upstream files the composition reads, by role, at their paths in
// upstream's tree. Every one is fetched on every run (about twenty-five small
// files), so what a setting selects never depends on what an earlier run
// downloaded.
const (
	fileNamespace          = "manifests/ate-install/ate-system-namespace.yaml"
	fileControllerRole     = "manifests/ate-install/generated/role.yaml"
	filePodCertificates    = "manifests/ate-install/pod-certificate-controller.yaml"
	fileSandboxValidation  = "manifests/ate-install/sandboxconfig-validation.yaml"
	fileSandboxGvisor      = "manifests/ate-install/sandboxconfig-gvisor.yaml"
	fileOtelConfig         = "manifests/ate-install/ate-otel-config.yaml"
	filePostgres           = "manifests/ate-install/postgres/postgres.yaml"
	filePostgresSize10     = "manifests/ate-install/postgres-size10/postgres-config-patch.yaml"
	filePodCertSize10      = "manifests/ate-install/podcert-size10/kustomization.yaml"
	fileAPIServer          = "manifests/ate-install/ate-api-server.yaml"
	fileController         = "manifests/ate-install/ate-controller.yaml"
	fileIdentityServer     = "manifests/ate-install/ate-idp-server.yaml"
	fileNodeAgent          = "manifests/ate-install/atelet.yaml"
	fileRouter             = "manifests/ate-install/atenet-router.yaml"
	fileRouterMonitoring   = "manifests/ate-install/atenet-router-monitoring.yaml"
	fileNetworkPolicies    = "manifests/ate-install/ate-system-networkpolicies.yaml"
	fileEgress             = "manifests/ate-install/atenet-egress.yaml"
	fileCloudSQLPatch      = "manifests/ate-install/cloudsql/proxy-sidecar-patch.yaml"
	fileAgentgateway       = "manifests/ate-install/components/agentgateway/kustomization.yaml"
	fileAgentgatewayConfig = "manifests/ate-install/components/agentgateway/configmap.yaml"
	fileAgentgatewayEgress = "manifests/ate-install/components/agentgateway-egress/kustomization.yaml"
	// The static-mTLS component patches from the file beside it (`path:`).
	fileAgentgatewayStaticMTLS       = "manifests/ate-install/components/agentgateway-static-mtls/kustomization.yaml"
	fileAgentgatewayStaticMTLSConfig = "manifests/ate-install/components/agentgateway-static-mtls/configmap.yaml"
	fileCordon                       = "manifests/ate-install/components/cordon-system-infra-components/kustomization.yaml"
	fileCredentialProvider           = "manifests/egress-credential-injection/k8s-credential-provider.yaml"
	fileSetupSQL                     = "pkg/postgressetup/setup.sql"
)

// releaseFiles is every upstream file the module reads besides the
// definitions (crdFiles).
var releaseFiles = []string{
	fileNamespace, fileControllerRole, filePodCertificates, fileSandboxValidation,
	fileSandboxGvisor, fileOtelConfig, filePostgres, filePostgresSize10, filePodCertSize10,
	fileAPIServer, fileController, fileIdentityServer, fileNodeAgent, fileRouter, fileRouterMonitoring,
	fileNetworkPolicies, fileEgress, fileCloudSQLPatch, fileAgentgateway, fileAgentgatewayConfig,
	fileAgentgatewayEgress, fileAgentgatewayStaticMTLS, fileAgentgatewayStaticMTLSConfig, fileCordon,
	fileCredentialProvider, fileSetupSQL,
}
