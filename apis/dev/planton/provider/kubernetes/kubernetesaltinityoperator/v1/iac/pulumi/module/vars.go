package module

var vars = struct {
	DefaultNamespace string
	HelmChartName    string
	HelmChartRepo    string
	HelmChartVersion string
}{
	DefaultNamespace: "kubernetes-altinity-operator",
	HelmChartName:    "altinity-clickhouse-operator",
	HelmChartRepo:    "https://docs.altinity.com/clickhouse-operator/",
	HelmChartVersion: "0.25.4",
}
