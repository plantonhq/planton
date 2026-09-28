package module

var vars = struct {
	// BarmanCloudPluginName is the CNPG-I plugin identifier the Cluster's
	// plugins list and every ScheduledBackup reference. It is the plugin's
	// registered name, fixed by upstream — not something users configure.
	BarmanCloudPluginName string

	// RecoverySourceExternalClusterName is the name the module gives the
	// synthetic externalClusters entry that recovery bootstraps read from.
	// Any stable label works (the entry only exists to carry the recovery
	// ObjectStore reference); "origin" follows upstream's own examples.
	RecoverySourceExternalClusterName string

	// BackupSeriesAnnotationKey carries the backup series on the Backup
	// every series starts from: a changed series changes the Backup, which
	// replaces it and so takes the new series' first base backup.
	BackupSeriesAnnotationKey string
}{
	BarmanCloudPluginName:             "barman-cloud.cloudnative-pg.io",
	RecoverySourceExternalClusterName: "origin",
	BackupSeriesAnnotationKey:         "planton.ai/backup-series",
}
