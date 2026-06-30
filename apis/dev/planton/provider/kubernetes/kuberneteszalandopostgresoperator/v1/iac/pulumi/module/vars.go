package module

// vars contain all tunable configurations for the Zalando Postgres‑Operator release.
var vars = struct {
	Namespace        string
	HelmChartName    string
	HelmChartRepo    string
	HelmChartVersion string
}{
	Namespace:        "postgres-operator",
	HelmChartName:    "postgres-operator",
	HelmChartRepo:    "https://opensource.zalando.com/postgres-operator/charts/postgres-operator",
	HelmChartVersion: "1.12.2",
}
