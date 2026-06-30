package module

// Output constants define the keys for stack outputs exported by the Helm Release deployment.
// These outputs are exported via Pulumi and can be retrieved using `pulumi stack output <key>`.
//
// The outputs correspond to the fields defined in KubernetesHelmReleaseStackOutputs proto message.
const (
	// OpNamespace is the Kubernetes namespace where the Helm release is deployed.
	// This namespace is created by the module and contains all resources from the Helm chart.
	//
	// Example usage:
	//   pulumi stack output namespace
	//
	// The namespace value is determined by (in priority order):
	//   1. Default: metadata.name from the KubernetesHelmRelease resource
	//   2. Override: custom label "planton.ai/kubernetes-namespace" if provided
	//   3. Override: kubernetes_namespace from KubernetesHelmReleaseStackInput if provided
	OpNamespace = "namespace"
)

// Output keys that may be added in future versions:
// - release_name: The name of the deployed Helm release
// - chart_name: The name of the Helm chart that was deployed
// - chart_version: The version of the Helm chart that was deployed
// - chart_repo: The repository URL from which the chart was fetched
// - release_status: The status of the Helm release (deployed, failed, etc.)
// - release_revision: The revision number of the Helm release
// - manifest: The rendered Kubernetes manifests from the Helm chart
//
// Note: These additional outputs would require corresponding fields to be added
// to the KubernetesHelmReleaseStackOutputs proto message first.
