package module

// vars carries the module's pinned chart identity. The chart version is
// the SERVED index truth (https://openbao.github.io/openbao-helm) — chart
// 0.28.6 pairs with OpenBao v2.6.1 (the chart's appVersion pins the
// server image tag).
var vars = struct {
	HelmChartName       string
	HelmChartRepo       string
	DefaultChartVersion string

	// OpenBao's API and cluster ports (chart constants — the listener
	// template binds [::]:8200 / [::]:8201).
	ApiPort     int
	ClusterPort int

	// Data/audit mount paths (chart constants — the config's storage
	// stanzas and the PVC mounts must agree).
	DataMountPath  string
	AuditMountPath string

	// Where the module mounts the TLS Secret when tls.enabled — the
	// listener's tls_cert_file/tls_key_file point here.
	TlsMountPath string

	// Suffix of the module-owned Secret carrying declared seal
	// credentials (AWS secret key / Azure client secret / transit
	// token), delivered to the server as environment variables so
	// nothing credential-bearing lands in the config ConfigMap.
	SealCredentialsSecretSuffix string

	// Name-length budgets, derived from the chart's name helpers: the
	// chart truncates its fullname at 63 but appends suffixes AFTER
	// truncation. Longest Service suffixes: `-internal` (9) always,
	// `-agent-injector-svc` (19) when the injector is enabled — Services
	// cap at 63 characters.
	MaxNameLength             int
	MaxNameLengthWithInjector int

	// Kubernetes caps CronJob names at 52 characters (the controller
	// appends `-<10-digit schedule time>` to each Job it creates and
	// does not truncate). `<name>-backup` (7) must fit, so declaring
	// `backup` lowers the budget to 45.
	MaxNameLengthWithBackup int

	// Suffixes of the module-owned backup objects, all hung off
	// metadata.name: the job's ServiceAccount (also the policy and the
	// default auth role — one name is the whole login recipe), the
	// scripts ConfigMap, the credentials Secret, the CronJob, and the
	// restore Job's prefix (its suffix hashes the declaration).
	BackupSuffix                  string
	BackupScriptsSuffix           string
	BackupCredentialsSecretSuffix string
	RestoreJobPrefix              string

	// The two job images. The `bao` CLI runs on the server's own image
	// at the chart's appVersion (tag without the `v`, as the chart
	// renders it) so the CLI never drifts from the server it snapshots;
	// rclone is the official image at the pin whose auth chains the
	// spec's keyless claims were read from. Both overridable through
	// `backup.images`.
	DefaultOpenBaoImage string
	DefaultRcloneImage  string

	// Where the job pods mount the scripts ConfigMap, the snapshot
	// scratch volume, and a declared S3 CA bundle.
	ScriptsMountPath   string
	SnapshotsMountPath string
	StoreCaMountPath   string

	// The job pods' identity: OpenBao's image runs as uid 100 (its
	// Dockerfile creates `openbao` with --uid 100); the rclone image
	// declares no user and would run as root. Both containers run under
	// this uid/gid so the 0600 snapshot file the bao CLI writes is
	// readable by the upload container without a chmod step.
	JobRunAsUser  int
	JobRunAsGroup int

	// Backup run hygiene: one run at a time, bounded retries, a deadline
	// sized for a large snapshot, bounded history. The restore Job
	// carries NO deadline — its wait for the operator's token Secret is
	// unbounded by design.
	BackupJobBackoffLimit          int
	BackupJobActiveDeadlineSeconds int
	BackupJobsHistoryLimit         int
	RestoreJobBackoffLimit         int

	// The client timeout the job's bao CLI runs with — snapshot calls
	// on large vaults outlive the CLI's 60s default.
	BaoClientTimeout string
}{
	HelmChartName:       "openbao",
	HelmChartRepo:       "https://openbao.github.io/openbao-helm",
	DefaultChartVersion: "0.28.6",

	ApiPort:     8200,
	ClusterPort: 8201,

	DataMountPath:  "/openbao/data",
	AuditMountPath: "/openbao/audit",
	TlsMountPath:   "/openbao/tls",

	SealCredentialsSecretSuffix: "-seal-credentials",

	MaxNameLength:             54,
	MaxNameLengthWithInjector: 44,
	MaxNameLengthWithBackup:   45,

	BackupSuffix:                  "-backup",
	BackupScriptsSuffix:           "-backup-scripts",
	BackupCredentialsSecretSuffix: "-backup-credentials",
	RestoreJobPrefix:              "-restore-",

	DefaultOpenBaoImage: "quay.io/openbao/openbao:2.6.1",
	DefaultRcloneImage:  "docker.io/rclone/rclone:1.75.1",

	ScriptsMountPath:   "/scripts",
	SnapshotsMountPath: "/snapshots",
	StoreCaMountPath:   "/etc/store-ca",

	JobRunAsUser:  100,
	JobRunAsGroup: 1000,

	BackupJobBackoffLimit:          2,
	BackupJobActiveDeadlineSeconds: 3600,
	BackupJobsHistoryLimit:         3,
	RestoreJobBackoffLimit:         3,

	BaoClientTimeout: "600s",
}
