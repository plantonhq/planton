package module

var vars = struct {
	HelmChartName        string
	HelmChartRepo        string
	CrdManifestURLFormat string
}{
	HelmChartName:        "solr-operator",
	HelmChartRepo:        "https://solr.apache.org/charts",
	CrdManifestURLFormat: "https://solr.apache.org/operator/downloads/crds/v%s/all-with-dependencies.yaml",
}
