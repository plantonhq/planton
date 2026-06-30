package module

// Helm chart constants for GHA Runner Scale Set.
var (
	// HelmChartOCI is the full OCI URL for the runner scale set chart.
	// For OCI charts, the full URL must be passed as the Chart parameter
	// (RepositoryOpts.Repo doesn't work with OCI registries in Pulumi).
	HelmChartOCI = "oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set"

	// DefaultChartVersion is the default version of the Helm chart.
	DefaultChartVersion = "0.13.1"
)

// Output keys for stack exports.
const (
	OpNamespace          = "namespace"
	OpReleaseName        = "release_name"
	OpChartVersion       = "chart_version"
	OpRunnerScaleSetName = "runner_scale_set_name"
	OpGitHubConfigURL    = "github_config_url"
	OpGitHubSecretName   = "github_secret_name"
	OpPvcNames           = "pvc_names"
	OpMinRunners         = "min_runners"
	OpMaxRunners         = "max_runners"
	OpContainerMode      = "container_mode"
)
