package module

var vars = struct {
	HelmChartName       string
	HelmChartRepo       string
	DefaultChartVersion string
	// ReleaseName is FIXED: Velero's CRDs and node-agent are
	// cluster-scoped and one server owns the backup records in the store —
	// one installation per cluster is an upstream constraint, so the
	// release name never derives from metadata.name. The fixed name also
	// pins the chart's derived names: velero.fullname collapses to
	// "velero" (release name contains the chart name), which is what makes
	// the service-account name below deterministic.
	ReleaseName string
	// ServerServiceAccountName is the chart-derived name of the Velero
	// server's ServiceAccount — the subject cloud-side keyless bindings
	// (IRSA trust policies, GCP WI bindings, Azure federated credentials)
	// are written against, so it is surfaced as an output.
	//
	// Derivation (chart templates/_helpers.tpl "velero.serverServiceAccount"):
	// serviceAccount.server.create defaults true and the module never sets
	// serviceAccount.server.name, so the name is
	// printf "%s-%s" (velero.fullname) "server". With the release named
	// "velero", velero.fullname is "velero" (the release name contains the
	// chart name), hence "velero-server".
	ServerServiceAccountName string
	// BackupStorageLocationName is the name the module gives the default
	// BackupStorageLocation — what Backup/Schedule resources reference
	// through storageLocation.
	BackupStorageLocationName string
	// Per-arm provider plugin defaults: the official images at the
	// versions paired with the chart's Velero release (chart 12.1.0 =
	// Velero 1.18). Overridable via spec.backup_storage.plugin_image for
	// private-registry mirrors or deliberate pins.
	AwsPluginImage   string
	GcpPluginImage   string
	AzurePluginImage string
	// Init-container names for each plugin (upstream convention shown in
	// the chart README's install example: velero-plugin-for-<PROVIDER>).
	AwsPluginName   string
	GcpPluginName   string
	AzurePluginName string
}{
	// Chart identity — MUST be identical in the Terraform module's locals
	// (cross-engine chart-name drift installs different software per
	// engine).
	HelmChartName: "velero",
	HelmChartRepo: "https://vmware-tanzu.github.io/helm-charts",
	// Fallback when spec.chart_version is unset AND the platform's
	// defaulting middleware did not run. Keep aligned with the spec
	// default and the chart-repo index (chart 12.1.0 ships Velero 1.18.1 —
	// chart and app versions move separately; the chart pin governs).
	DefaultChartVersion:       "12.1.0",
	ReleaseName:               "velero",
	ServerServiceAccountName:  "velero-server",
	BackupStorageLocationName: "default",
	AwsPluginImage:            "velero/velero-plugin-for-aws:v1.14.2",
	GcpPluginImage:            "velero/velero-plugin-for-gcp:v1.14.2",
	AzurePluginImage:          "velero/velero-plugin-for-microsoft-azure:v1.14.2",
	AwsPluginName:             "velero-plugin-for-aws",
	GcpPluginName:             "velero-plugin-for-gcp",
	AzurePluginName:           "velero-plugin-for-azure",
}
