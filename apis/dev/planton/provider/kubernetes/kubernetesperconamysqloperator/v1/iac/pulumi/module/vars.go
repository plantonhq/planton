package module

var vars = struct {
	HelmChartName    string
	HelmChartRepo    string
	HelmChartVersion string
}{
	HelmChartName:    "pxc-operator",
	HelmChartRepo:    "https://percona.github.io/percona-helm-charts/",
	HelmChartVersion: "1.18.0",
}
