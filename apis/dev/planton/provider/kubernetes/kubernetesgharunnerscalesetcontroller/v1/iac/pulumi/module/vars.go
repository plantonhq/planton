package module

// vars groups all operator-level constants so version bumps or repo changes
// remain one-liner edits.
var vars = struct {
	// Helm chart configuration
	// For OCI charts, the full URL must be passed as the Chart parameter
	// (RepositoryOpts.Repo doesn't work with OCI registries in Pulumi)
	HelmChartOCI string

	// Default Helm chart version (can be overridden via spec)
	DefaultChartVersion string
}{
	HelmChartOCI:        "oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set-controller",
	DefaultChartVersion: "0.13.1",
}
