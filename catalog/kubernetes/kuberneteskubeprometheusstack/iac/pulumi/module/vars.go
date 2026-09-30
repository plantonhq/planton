package module

// Chart identity — must stay byte-identical with the Terraform module's
// locals (helm_chart_name / helm_chart_repo): cross-engine chart-name drift
// deploys two different products from one manifest.
var vars = struct {
	HelmChartName string
	HelmChartRepo string
	// Fallback when spec.chart_version is unset — mirror of the proto
	// field's default option and the Terraform module's coalesce. Chart
	// 91.8.2 pairs with Prometheus Operator v0.94.1; the chart pin
	// governs.
	DefaultChartVersion string
	// Fallbacks for the storage arms — mirrors of the proto defaults.
	DefaultPrometheusDiskSize   string
	DefaultAlertmanagerDiskSize string
	DefaultGrafanaStorageSize   string
	// The chart SILENTLY truncates fullnameOverride at this many
	// characters (its own headroom for the longest child name it
	// derives). The modules pin the fullname to the resource name and
	// FAIL LOUDLY on longer names instead of letting the chart truncate —
	// a truncated fullname breaks the naming contract every exported
	// output is built on.
	FullnameBudget int
	// Suffix key names of the module-owned Secret that carries declared
	// remote-write basic-auth usernames. The Prometheus CRD reads BOTH
	// basic-auth halves from Secrets; usernames are not secrets, so the
	// spec accepts them as plain strings and the module materializes this
	// Secret (the declared-credentials pattern) rather than pushing a
	// pre-created Secret onto the user for a non-secret value.
	RemoteWriteAuthSecretSuffix string
	// Suffix of the module-owned Secret carrying the notification
	// credentials (spec.alertmanager.notifications): every credential is a
	// managed-secret reference the platform resolves at deploy, and the
	// module writes the resolved values here for Alertmanager's `_file`
	// fields to read.
	NotificationsSecretSuffix string
	// Fallbacks mirroring the notifications proto defaults: the heartbeat
	// interval and the labels that title every message.
	DefaultHeartbeatInterval string
	DefaultEnvironmentLabel  string
	DefaultComponentLabel    string
}{
	HelmChartName:               "kube-prometheus-stack",
	HelmChartRepo:               "https://prometheus-community.github.io/helm-charts",
	DefaultChartVersion:         "91.8.2",
	DefaultPrometheusDiskSize:   "50Gi",
	DefaultAlertmanagerDiskSize: "2Gi",
	DefaultGrafanaStorageSize:   "10Gi",
	FullnameBudget:              26,
	RemoteWriteAuthSecretSuffix: "-remote-write-auth",
	NotificationsSecretSuffix:   "-alertmanager-notifications",
	DefaultHeartbeatInterval:    "1m",
	DefaultEnvironmentLabel:     "environment",
	DefaultComponentLabel:       "component",
}
