package module

// vars groups all operator‑level constants so version bumps or repo changes
// remain one‑liner edits.
var vars = struct {
	Namespace        string
	HelmChartName    string
	HelmChartRepo    string
	HelmChartVersion string
}{
	Namespace:        "elastic-system",
	HelmChartName:    "eck-operator",
	HelmChartRepo:    "https://helm.elastic.co",
	HelmChartVersion: "2.14.0", // update when you move the stable channel
}
