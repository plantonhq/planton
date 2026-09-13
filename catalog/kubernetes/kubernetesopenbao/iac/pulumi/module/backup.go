package module

import (
	"encoding/base64"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesopenbaov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopenbao/v1alpha1"
	"github.com/plantonhq/planton/pkg/cloudflare/r2"
	batchv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/batch/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The rclone remote the jobs address is named `store`; every option
// reaches rclone as RCLONE_CONFIG_STORE_<OPTION> — no config file with
// credentials is ever rendered. Non-secret options (type, provider,
// region, endpoint, addressing) ride the container's plain env; the
// secret half (keys, service-account JSON, connection strings) rides the
// module-owned `<name>-backup-credentials` Secret under the SAME variable
// names, wired in as secretKeyRef entries.
const rcloneRemotePrefix = "RCLONE_CONFIG_STORE_"

// The kind of store, for the scripts' permission hints.
const (
	storeKindS3    = "s3"
	storeKindGcs   = "gcs"
	storeKindAzure = "azure"
	storeKindR2    = "r2"
)

// Volume and mount names shared by the CronJob and the restore Job.
const (
	volScripts   = "scripts"
	volSnapshots = "snapshots"
	volTls       = "tls"
	volStoreCa   = "store-ca"
	storeCaKey   = "ca.pem"
)

// workloadIdentityAnnotations translates the shared workload-identity
// oneof into the exact ServiceAccount annotations each cloud's webhook
// expects — the same mapping every catalog consumer of the message uses
// (Postgres, cert-manager, external-dns, ESO, ServiceAccount).
func workloadIdentityAnnotations(wi *kubernetesprovider.KubernetesWorkloadIdentity) map[string]string {
	if wi == nil {
		return nil
	}
	annotations := map[string]string{}
	if gke := wi.GetGke(); gke != nil {
		annotations["iam.gke.io/gcp-service-account"] = gke.GetServiceAccountEmail().GetValue()
	}
	if eks := wi.GetEks(); eks != nil {
		annotations["eks.amazonaws.com/role-arn"] = eks.GetRoleArn().GetValue()
	}
	if aks := wi.GetAks(); aks != nil {
		annotations["azure.workload.identity/client-id"] = aks.GetClientId().GetValue()
		if aks.TenantId != nil && aks.GetTenantId() != "" {
			annotations["azure.workload.identity/tenant-id"] = aks.GetTenantId()
		}
	}
	return annotations
}

// storeKind names the declared arm.
func storeKind(store *kubernetesopenbaov1alpha1.KubernetesOpenBaoBackupObjectStore) string {
	switch {
	case store.GetR2() != nil:
		return storeKindR2
	case store.GetGcs() != nil:
		return storeKindGcs
	case store.GetAzureBlob() != nil:
		return storeKindAzure
	default:
		return storeKindS3
	}
}

// storePath is the rclone path the jobs read and write: `store:<bucket or
// container>/<prefix>` (no trailing slash; the scripts append the object
// name). storeRoot is the same without the prefix, for a restore that
// names its key relative to the bucket.
func storePaths(store *kubernetesopenbaov1alpha1.KubernetesOpenBaoBackupObjectStore) (root, path string) {
	var bucket string
	switch {
	case store.GetR2() != nil:
		bucket = store.GetR2().GetBucket().GetValue()
	case store.GetGcs() != nil:
		bucket = store.GetGcs().GetBucket().GetValue()
	case store.GetAzureBlob() != nil:
		bucket = store.GetAzureBlob().GetContainer()
	default:
		bucket = store.GetS3().GetBucket()
	}
	root = "store:" + bucket
	path = root
	if prefix := strings.Trim(store.GetPrefix(), "/"); prefix != "" {
		path = root + "/" + prefix
	}
	return root, path
}

// storePlainEnv returns the NON-secret rclone remote configuration for
// the declared arm, keyed by the full RCLONE_CONFIG_STORE_* variable.
// Twin: local.backup_store_plain_env in locals.tf.
//
// The S3 dialect each store needs is rclone's business, selected by
// `provider`: Cloudflare (path-style, no multipart ETags) for R2, AWS for
// real S3, and rclone's conservative `Other` quirk set for any declared
// endpoint. `no_check_bucket` on every arm: the bucket is declared to
// exist (by reference to the catalog's bucket kind), and a bucket-scoped
// credential is refused the HeadBucket/CreateBucket rclone would
// otherwise attempt. GCS gets `bucket_policy_only` because a bucket with
// uniform bucket-level access refuses the per-object ACL writes rclone
// would otherwise make.
func storePlainEnv(store *kubernetesopenbaov1alpha1.KubernetesOpenBaoBackupObjectStore) map[string]string {
	env := map[string]string{}
	set := func(option, value string) { env[rcloneRemotePrefix+strings.ToUpper(option)] = value }
	switch {
	case store.GetR2() != nil:
		r2Store := store.GetR2()
		set("type", "s3")
		set("provider", "Cloudflare")
		set("region", r2.Region)
		set("endpoint", r2.S3Endpoint(r2Store.GetAccountId().GetValue(), r2Store.GetJurisdiction().GetValue()))
		set("force_path_style", "true")
		set("no_check_bucket", "true")
	case store.GetGcs() != nil:
		set("type", "google cloud storage")
		set("bucket_policy_only", "true")
		set("no_check_bucket", "true")
		if store.GetGcs().GetKeyless() {
			set("env_auth", "true")
		}
	case store.GetAzureBlob() != nil:
		az := store.GetAzureBlob()
		set("type", "azureblob")
		set("no_check_container", "true")
		if az.GetStorageAccount() != "" {
			set("account", az.GetStorageAccount())
		}
		if az.GetKeyless() {
			set("env_auth", "true")
		}
	default:
		s3 := store.GetS3()
		set("type", "s3")
		set("no_check_bucket", "true")
		if endpoint := s3.GetEndpointUrl().GetValue(); endpoint != "" {
			set("provider", "Other")
			set("endpoint", endpoint)
		} else {
			set("provider", "AWS")
		}
		if s3.GetRegion() != "" {
			set("region", s3.GetRegion())
		}
		if s3.GetForcePathStyle() {
			set("force_path_style", "true")
		}
		if s3.GetKeyless() {
			set("env_auth", "true")
		}
	}
	return env
}

// backupCredentialsSecretData returns the SECRET half of the remote
// configuration, keyed by the same RCLONE_CONFIG_STORE_* variables, plus
// the `ca.pem` file when the s3 arm declares one — or an empty map for
// the keyless postures, which need no Secret. Twin:
// local.backup_credentials_data in locals.tf.
func backupCredentialsSecretData(backup *kubernetesopenbaov1alpha1.KubernetesOpenBaoBackup) (map[string]string, error) {
	data := map[string]string{}
	set := func(option, value string) { data[rcloneRemotePrefix+strings.ToUpper(option)] = value }
	store := backup.GetObjectStore()
	switch {
	case store.GetR2() != nil:
		creds := store.GetR2().GetCredentials()
		set("access_key_id", creds.GetAccessKeyId().GetValue())
		set("secret_access_key", creds.GetSecretAccessKey().GetValue())
	case store.GetGcs() != nil:
		if key := store.GetGcs().GetServiceAccountKey().GetValue(); key != "" {
			decoded, err := decodeServiceAccountKey(key)
			if err != nil {
				return nil, err
			}
			set("service_account_credentials", decoded)
		}
	case store.GetAzureBlob() != nil:
		az := store.GetAzureBlob()
		if az.GetStorageKey() != "" {
			set("key", az.GetStorageKey())
		}
		if az.GetConnectionString() != "" {
			set("connection_string", az.GetConnectionString())
		}
	default:
		s3 := store.GetS3()
		if keys := s3.GetAccessKeys(); keys != nil {
			set("access_key_id", keys.GetAccessKeyId())
			set("secret_access_key", keys.GetSecretAccessKey())
		}
		if s3.GetCaPem() != "" {
			data[storeCaKey] = s3.GetCaPem()
		}
	}
	return data, nil
}

// decodeServiceAccountKey accepts the key file two ways — raw JSON, or the
// base64 encoding a GcpServiceAccount resource exports as `key_base64` —
// and returns the JSON text rclone reads. Raw JSON is recognized by its
// opening brace; anything else must be valid standard base64. Twin:
// local.backup_gcs_key_json in locals.tf.
func decodeServiceAccountKey(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "{") {
		return trimmed, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return "", errors.Wrap(err, "the GCS service account key is neither a JSON key file nor its base64 encoding (the GcpServiceAccount key_base64 output)")
	}
	return string(decoded), nil
}

// sortedKeys returns a map's keys in order — env lists and Secret
// wiring must render deterministically or every apply diffs.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// plainEnv renders a map as container env entries, sorted.
func plainEnv(m map[string]string) kubernetescorev1.EnvVarArray {
	out := kubernetescorev1.EnvVarArray{}
	for _, k := range sortedKeys(m) {
		out = append(out, &kubernetescorev1.EnvVarArgs{Name: pulumi.String(k), Value: pulumi.String(m[k])})
	}
	return out
}

// secretEnv renders one secretKeyRef entry per credential key (the
// `ca.pem` file is mounted, not exported).
func secretEnv(secretName string, data map[string]string) kubernetescorev1.EnvVarArray {
	out := kubernetescorev1.EnvVarArray{}
	for _, k := range sortedKeys(data) {
		if k == storeCaKey {
			continue
		}
		out = append(out, &kubernetescorev1.EnvVarArgs{
			Name: pulumi.String(k),
			ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
				SecretKeyRef: &kubernetescorev1.SecretKeySelectorArgs{
					Name: pulumi.String(secretName),
					Key:  pulumi.String(k),
				},
			},
		})
	}
	return out
}

// jobPodSecurityContext is the identity both job containers run under
// (see vars.JobRunAsUser).
func jobPodSecurityContext() *kubernetescorev1.PodSecurityContextArgs {
	return &kubernetescorev1.PodSecurityContextArgs{
		RunAsUser:    pulumi.Int(vars.JobRunAsUser),
		RunAsGroup:   pulumi.Int(vars.JobRunAsGroup),
		FsGroup:      pulumi.Int(vars.JobRunAsGroup),
		RunAsNonRoot: pulumi.Bool(true),
		SeccompProfile: &kubernetescorev1.SeccompProfileArgs{
			Type: pulumi.String("RuntimeDefault"),
		},
	}
}

func jobContainerSecurityContext() *kubernetescorev1.SecurityContextArgs {
	return &kubernetescorev1.SecurityContextArgs{
		AllowPrivilegeEscalation: pulumi.Bool(false),
		Capabilities: &kubernetescorev1.CapabilitiesArgs{
			Drop: pulumi.StringArray{pulumi.String("ALL")},
		},
	}
}

// baoClientEnv is what the bao CLI containers need to reach the server:
// the active-leader address, the CA when TLS is on, and a client timeout
// sized for large snapshots. Never an empty BAO_* value — an empty BAO_
// variable overrides its VAULT_ counterpart in the client.
func baoClientEnv(locals *Locals) map[string]string {
	env := map[string]string{
		"BAO_ADDR":           locals.BaoAddr,
		"BAO_CLIENT_TIMEOUT": vars.BaoClientTimeout,
		"SNAPSHOT_DIR":       vars.SnapshotsMountPath,
	}
	if locals.TlsEnabled {
		env["BAO_CACERT"] = vars.TlsMountPath + "/ca.crt"
	}
	return env
}

// jobVolumes are the pod volumes both jobs share: the scripts, the
// snapshot scratch space, the server's TLS Secret when TLS is on (for
// the CA), and the declared store CA when the s3 arm carries one.
func jobVolumes(locals *Locals, credentials map[string]string) kubernetescorev1.VolumeArray {
	volumes := kubernetescorev1.VolumeArray{
		&kubernetescorev1.VolumeArgs{
			Name: pulumi.String(volScripts),
			ConfigMap: &kubernetescorev1.ConfigMapVolumeSourceArgs{
				Name:        pulumi.String(locals.BackupScriptsName),
				DefaultMode: pulumi.Int(0o555),
			},
		},
		&kubernetescorev1.VolumeArgs{
			Name:     pulumi.String(volSnapshots),
			EmptyDir: &kubernetescorev1.EmptyDirVolumeSourceArgs{},
		},
	}
	if locals.TlsEnabled {
		volumes = append(volumes, &kubernetescorev1.VolumeArgs{
			Name: pulumi.String(volTls),
			Secret: &kubernetescorev1.SecretVolumeSourceArgs{
				SecretName: pulumi.String(locals.TlsSecretName),
			},
		})
	}
	if _, ok := credentials[storeCaKey]; ok {
		volumes = append(volumes, &kubernetescorev1.VolumeArgs{
			Name: pulumi.String(volStoreCa),
			Secret: &kubernetescorev1.SecretVolumeSourceArgs{
				SecretName: pulumi.String(locals.BackupCredentialsSecretName),
				Items: kubernetescorev1.KeyToPathArray{
					&kubernetescorev1.KeyToPathArgs{Key: pulumi.String(storeCaKey), Path: pulumi.String(storeCaKey)},
				},
			},
		})
	}
	return volumes
}

// baoContainerMounts: scripts, snapshots, and the TLS CA.
func baoContainerMounts(locals *Locals) kubernetescorev1.VolumeMountArray {
	mounts := kubernetescorev1.VolumeMountArray{
		&kubernetescorev1.VolumeMountArgs{Name: pulumi.String(volScripts), MountPath: pulumi.String(vars.ScriptsMountPath), ReadOnly: pulumi.Bool(true)},
		&kubernetescorev1.VolumeMountArgs{Name: pulumi.String(volSnapshots), MountPath: pulumi.String(vars.SnapshotsMountPath)},
	}
	if locals.TlsEnabled {
		mounts = append(mounts, &kubernetescorev1.VolumeMountArgs{Name: pulumi.String(volTls), MountPath: pulumi.String(vars.TlsMountPath), ReadOnly: pulumi.Bool(true)})
	}
	return mounts
}

// rcloneContainerMounts: scripts, snapshots, and the store CA.
func rcloneContainerMounts(credentials map[string]string) kubernetescorev1.VolumeMountArray {
	mounts := kubernetescorev1.VolumeMountArray{
		&kubernetescorev1.VolumeMountArgs{Name: pulumi.String(volScripts), MountPath: pulumi.String(vars.ScriptsMountPath), ReadOnly: pulumi.Bool(true)},
		&kubernetescorev1.VolumeMountArgs{Name: pulumi.String(volSnapshots), MountPath: pulumi.String(vars.SnapshotsMountPath)},
	}
	if _, ok := credentials[storeCaKey]; ok {
		mounts = append(mounts, &kubernetescorev1.VolumeMountArgs{Name: pulumi.String(volStoreCa), MountPath: pulumi.String(vars.StoreCaMountPath), ReadOnly: pulumi.Bool(true)})
	}
	return mounts
}

// rcloneEnv is the full environment of an rclone container: the plain
// remote options, the secret half from the Secret, the store paths, and
// the CA path when declared.
func rcloneEnv(locals *Locals, extra map[string]string) kubernetescorev1.EnvVarArray {
	store := locals.Spec.GetBackup().GetObjectStore()
	root, path := storePaths(store)
	plain := storePlainEnv(store)
	plain["STORE_ROOT"] = root
	plain["STORE_PATH"] = path
	plain["STORE_KIND"] = storeKind(store)
	plain["SNAPSHOT_DIR"] = vars.SnapshotsMountPath
	credentials, _ := backupCredentialsSecretData(locals.Spec.GetBackup())
	if _, ok := credentials[storeCaKey]; ok {
		plain["RCLONE_CA_CERT"] = vars.StoreCaMountPath + "/" + storeCaKey
	}
	for k, v := range extra {
		plain[k] = v
	}
	env := plainEnv(plain)
	if locals.BackupCredentialsSecretName != "" {
		env = append(env, secretEnv(locals.BackupCredentialsSecretName, credentials)...)
	}
	return env
}

// jobPodLabels are the pod-template labels: the module's identity labels
// plus the AKS workload-identity label when the job federates with an
// Azure identity (the webhook injects the token only into labeled pods;
// the shared identity proto documents that this half lives on the
// workload).
func jobPodLabels(locals *Locals) map[string]string {
	labels := map[string]string{}
	for k, v := range locals.Labels {
		labels[k] = v
	}
	if locals.Spec.GetBackup().GetWorkloadIdentity().GetAks() != nil {
		labels["azure.workload.identity/use"] = "true"
	}
	return labels
}

// backupResources renders the ServiceAccount, the scripts ConfigMap, the
// credentials Secret (when an arm declares keys), and the CronJob. The
// release does not wait on any of them; the CronJob's first run fails its
// login until the operator runs the recipe — by design, and taught by the
// run's own log.
func backupResources(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.Resource) ([]pulumi.Resource, error) {
	backup := locals.Spec.GetBackup()
	opts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	if len(dependsOn) > 0 {
		opts = append(opts, pulumi.DependsOn(dependsOn))
	}

	// The job's ServiceAccount. Its annotation is the cloud-side half of
	// the keyless posture; the other half is the binding the identity
	// kind owns (a GcpGkeWorkloadIdentityBinding naming this SA).
	saMeta := &kubernetesmeta.ObjectMetaArgs{
		Name:      pulumi.String(locals.BackupName),
		Namespace: pulumi.String(locals.Namespace),
		Labels:    pulumi.ToStringMap(locals.Labels),
	}
	if annotations := workloadIdentityAnnotations(backup.GetWorkloadIdentity()); len(annotations) > 0 {
		saMeta.Annotations = pulumi.ToStringMap(annotations)
	}
	createdSa, err := kubernetescorev1.NewServiceAccount(ctx, locals.BackupName,
		&kubernetescorev1.ServiceAccountArgs{Metadata: kubernetesmeta.ObjectMetaPtrInput(saMeta)}, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create backup service account")
	}

	createdScripts, err := kubernetescorev1.NewConfigMap(ctx, locals.BackupScriptsName,
		&kubernetescorev1.ConfigMapArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.BackupScriptsName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			}),
			Data: pulumi.ToStringMap(backupScripts()),
		}, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create backup scripts configmap")
	}

	jobDeps := []pulumi.Resource{createdSa, createdScripts}

	credentials, err := backupCredentialsSecretData(backup)
	if err != nil {
		return nil, err
	}
	if locals.BackupCredentialsSecretName != "" {
		createdSecret, err := kubernetescorev1.NewSecret(ctx, locals.BackupCredentialsSecretName,
			&kubernetescorev1.SecretArgs{
				Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
					Name:      pulumi.String(locals.BackupCredentialsSecretName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				}),
				StringData: pulumi.ToStringMap(credentials),
			}, opts...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create backup credentials secret")
		}
		jobDeps = append(jobDeps, createdSecret)
	}

	if err := backupCronJob(ctx, locals, kubernetesProvider, jobDeps, credentials); err != nil {
		return nil, err
	}
	return jobDeps, nil
}

// backupCronJob renders the scheduled snapshot: an init container takes
// the snapshot with the bao CLI, the main container ships and prunes with
// rclone. One run at a time; a run that outlives the deadline is a
// failure, not a wait. SUSPENDED while `restore` is declared (see the
// spec's restore field and locals.RestoreDeclared).
func backupCronJob(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.Resource, credentials map[string]string) error {
	backup := locals.Spec.GetBackup()

	schedule := "0 * * * *"
	if backup.Schedule != nil && backup.GetSchedule() != "" {
		schedule = backup.GetSchedule()
	}
	retentionDays := 14
	if backup.RetentionDays != nil {
		retentionDays = int(backup.GetRetentionDays())
	}

	snapshotEnv := baoClientEnv(locals)
	snapshotEnv["BAO_AUTH_PATH"] = locals.BackupAuthMountPath
	snapshotEnv["BAO_ROLE"] = locals.BackupAuthRole
	snapshotEnv["BACKUP_POLICY"] = locals.BackupName
	snapshotEnv["BACKUP_SERVICE_ACCOUNT"] = locals.BackupName
	snapshotEnv["BACKUP_NAMESPACE"] = locals.Namespace
	snapshotEnv["SNAPSHOT_PREFIX"] = locals.ReleaseName

	resources := resourcesBlock(backup.GetResources())

	podSpec := &kubernetescorev1.PodSpecArgs{
		ServiceAccountName: pulumi.String(locals.BackupName),
		RestartPolicy:      pulumi.String("OnFailure"),
		SecurityContext:    jobPodSecurityContext(),
		InitContainers: kubernetescorev1.ContainerArray{
			&kubernetescorev1.ContainerArgs{
				Name:            pulumi.String("snapshot"),
				Image:           pulumi.String(locals.OpenBaoImage),
				Command:         pulumi.ToStringArray([]string{"sh", vars.ScriptsMountPath + "/" + scriptSnapshot}),
				Env:             plainEnv(snapshotEnv),
				VolumeMounts:    baoContainerMounts(locals),
				SecurityContext: jobContainerSecurityContext(),
				Resources:       resourceRequirements(resources),
			},
		},
		Containers: kubernetescorev1.ContainerArray{
			&kubernetescorev1.ContainerArgs{
				Name:    pulumi.String("upload"),
				Image:   pulumi.String(locals.RcloneImage),
				Command: pulumi.ToStringArray([]string{"sh", vars.ScriptsMountPath + "/" + scriptUpload}),
				Env: rcloneEnv(locals, map[string]string{
					"RETENTION_DAYS": strconv.Itoa(retentionDays),
				}),
				VolumeMounts:    rcloneContainerMounts(credentials),
				SecurityContext: jobContainerSecurityContext(),
				Resources:       resourceRequirements(resources),
			},
		},
		Volumes: jobVolumes(locals, credentials),
	}
	if sched := locals.Spec.GetServer().GetScheduling(); sched != nil {
		if len(sched.GetNodeSelector()) > 0 {
			podSpec.NodeSelector = pulumi.ToStringMap(sched.GetNodeSelector())
		}
		if len(sched.GetTolerations()) > 0 {
			podSpec.Tolerations = tolerationArray(sched.GetTolerations())
		}
	}

	_, err := batchv1.NewCronJob(ctx, locals.BackupName,
		&batchv1.CronJobArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.BackupName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			}),
			Spec: &batchv1.CronJobSpecArgs{
				Schedule:          pulumi.String(schedule),
				ConcurrencyPolicy: pulumi.String("Forbid"),
				// RESTORE MODE: a fresh target must not snapshot an empty
				// vault into the shared prefix (and prune the source's
				// snapshots) while its restore Job waits for the token.
				Suspend:                    pulumi.Bool(locals.RestoreDeclared),
				SuccessfulJobsHistoryLimit: pulumi.Int(vars.BackupJobsHistoryLimit),
				FailedJobsHistoryLimit:     pulumi.Int(vars.BackupJobsHistoryLimit),
				JobTemplate: &batchv1.JobTemplateSpecArgs{
					Spec: &batchv1.JobSpecArgs{
						BackoffLimit:          pulumi.Int(vars.BackupJobBackoffLimit),
						ActiveDeadlineSeconds: pulumi.Int(vars.BackupJobActiveDeadlineSeconds),
						Template: &kubernetescorev1.PodTemplateSpecArgs{
							Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
								Labels: pulumi.ToStringMap(jobPodLabels(locals)),
							}),
							Spec: podSpec,
						},
					},
				},
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, pulumi.DependsOn(dependsOn))...)
	if err != nil {
		return errors.Wrap(err, "failed to create backup cronjob")
	}
	return nil
}

// resourceRequirements renders the chart-shaped resources map (from
// resourcesBlock) as a core/v1 ResourceRequirements; nil stays nil.
func resourceRequirements(block map[string]interface{}) kubernetescorev1.ResourceRequirementsPtrInput {
	if block == nil {
		return nil
	}
	args := &kubernetescorev1.ResourceRequirementsArgs{}
	if requests, ok := block["requests"].(map[string]interface{}); ok {
		args.Requests = pulumi.ToStringMap(interfaceMapToString(requests))
	}
	if limits, ok := block["limits"].(map[string]interface{}); ok {
		args.Limits = pulumi.ToStringMap(interfaceMapToString(limits))
	}
	return args
}

func interfaceMapToString(in map[string]interface{}) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

// tolerationArray renders the shared WorkloadToleration list for a pod
// spec (the jobs inherit server.scheduling, so they land where the
// servers may).
func tolerationArray(tolerations []*kubernetesprovider.WorkloadToleration) kubernetescorev1.TolerationArray {
	out := kubernetescorev1.TolerationArray{}
	for _, t := range tolerations {
		tol := &kubernetescorev1.TolerationArgs{}
		if t.GetKey() != "" {
			tol.Key = pulumi.String(t.GetKey())
		}
		if t.GetOperator() != "" {
			tol.Operator = pulumi.String(t.GetOperator())
		}
		if t.GetValue() != "" {
			tol.Value = pulumi.String(t.GetValue())
		}
		if t.GetEffect() != "" {
			tol.Effect = pulumi.String(t.GetEffect())
		}
		if t.TolerationSeconds != nil {
			tol.TolerationSeconds = pulumi.Int(int(t.GetTolerationSeconds()))
		}
		out = append(out, tol)
	}
	return out
}
