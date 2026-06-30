package module

var vars = struct {
	// Namespace is the default namespace if not specified in spec
	Namespace string
	// HelmChartName is the actual Helm chart name (not the release name)
	HelmChartName string
	// HelmChartRepo is the Helm repository URL
	HelmChartRepo string
	// DefaultStableVersion is the default chart version
	DefaultStableVersion string
	// DefaultSecretsPollIntervalSec is the polling interval for secrets
	// excessive polling can become expensive on some providers
	DefaultSecretsPollIntervalSec int
}{
	Namespace:                     "kubernetes-external-secrets",
	HelmChartName:                 "external-secrets",
	HelmChartRepo:                 "https://charts.external-secrets.io",
	DefaultStableVersion:          "0.9.20", // bump when you move the stable channel
	DefaultSecretsPollIntervalSec: 10,
}
