package module

// Central place to keep chart / namespace constants.
// (Helps maintain consistency between main.go and the helper that creates the
// Helm release.)
var vars = struct {
	Namespace        string
	HelmChartName    string
	HelmChartRepo    string
	HelmChartVersion string
}{
	Namespace:     "strimzi-operator",
	HelmChartName: "strimzi-kafka-operator",
	// https://artifacthub.io/packages/helm/strimzi/strimzi-kafka-operator
	HelmChartRepo:    "https://strimzi.io/charts/",
	HelmChartVersion: "0.42.0",
}
