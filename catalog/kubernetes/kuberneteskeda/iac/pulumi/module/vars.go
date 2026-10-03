package module

var vars = struct {
	HelmChartName       string
	HelmChartRepo       string
	DefaultChartVersion string
	// ReleaseName is FIXED: KEDA registers the cluster-wide
	// v1beta1.external.metrics.k8s.io APIService, a singleton — one
	// installation per cluster is an upstream constraint, so the release
	// name never derives from metadata.name.
	ReleaseName string
	// OperatorServiceAccountName is the chart's FIXED service-account name
	// for the operator ("keda-operator" regardless of release name) — the
	// subject cloud-side keyless bindings (IRSA trust policies, GCP WI
	// bindings, Entra federated credentials) are written against, so it is
	// surfaced as an output.
	OperatorServiceAccountName string
}{
	// Chart identity — MUST be identical in the Terraform module's locals
	// (cross-engine chart-name drift installs different software per
	// engine).
	HelmChartName: "keda",
	HelmChartRepo: "https://kedacore.github.io/charts",
	// Fallback when spec.chart_version is unset AND the platform's
	// defaulting middleware did not run. Keep aligned with the spec
	// default and the chart-repo index (chart 2.20.1 = KEDA 2.20.1 —
	// chart and app versions move together).
	DefaultChartVersion:        "2.20.1",
	ReleaseName:                "keda",
	OperatorServiceAccountName: "keda-operator",
}
