package module

var vars = struct {
	Namespace           string
	HelmChartName       string
	HelmChartRepo       string
	DefaultChartVersion string
}{
	Namespace:           "kubernetes-ingress-nginx",
	HelmChartName:       "ingress-nginx",
	HelmChartRepo:       "https://kubernetes.github.io/ingress-nginx",
	DefaultChartVersion: "4.11.1", // bump when you move the stable channel
}
