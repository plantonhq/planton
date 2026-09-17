package module

var vars = struct {
	// HelmOciRepo is the OCI registry path holding the planton-runner
	// chart. Pulumi's helm.v3.Release does not resolve oci:// through
	// RepositoryOpts the way the Terraform provider does — the chart
	// reference must be the JOINED "<repo>/<chart>" string (see main.go).
	HelmOciRepo string

	// HelmChartName is the official runner chart.
	HelmChartName string

	// DefaultChartVersion is the chart this catalog release was validated
	// against — the version installed when spec.chart_version is unset.
	// Bump it together with a re-validation of the values contract below
	// (enrollment block, resources, build block). 0.5.0 probes the pod over
	// gRPC on the runner's port, the contract every execution mode serves;
	// 0.4.0 probed the tunnel agent's HTTP port and restarted a runner
	// enrolled with a tunnel-less instance every liveness window.
	DefaultChartVersion string

	// MinChartVersion is the enrollment-contract floor: charts below
	// 0.4.0 predate token enrollment and silently IGNORE the enrollment
	// values — the runner would deploy with no way to join. Refused
	// loudly in both engines instead.
	MinChartVersion string

	// TokenSecretSuffix names the module-created Secret (`<name>-token`)
	// the chart reads the runner token from (its existingSecret form) —
	// the token never rides rendered chart values.
	TokenSecretSuffix string

	// TokenSecretKey is the key inside the Secret. Matches the chart's
	// own default so the values only need the Secret NAME.
	TokenSecretKey string
}{
	HelmOciRepo:         "oci://ghcr.io/plantonhq/charts",
	HelmChartName:       "planton-runner",
	DefaultChartVersion: "0.5.0",
	MinChartVersion:     "0.4.0",
	TokenSecretSuffix:   "-token",
	TokenSecretKey:      "token",
}
