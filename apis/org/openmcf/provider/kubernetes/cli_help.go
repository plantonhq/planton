package kubernetes

// CLI help constants for Kubernetes provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the Kubernetes provider.
// These are read by the Pulumi/Terraform Kubernetes providers when no explicit config file is provided.
var EnvironmentVariables = []string{
	"KUBECONFIG",
	"KUBE_CONFIG_PATH",
	"KUBE_CONTEXT",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export KUBECONFIG="/path/to/kubeconfig"
# Or use default: ~/.kube/config`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `provider: 1  # 1=GCP_GKE, 2=AWS_EKS, 3=AZURE_AKS, 4=DIGITAL_OCEAN_DOKS
gcp_gke:
  cluster_endpoint: "<cluster-endpoint>"
  cluster_ca_data: "<base64-encoded-ca-cert>"
  service_account_key: "<service-account-json>"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "kubernetes-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "Kubernetes"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/"
