package module

import (
	"github.com/pkg/errors"
	kubernetespostgresv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespostgres/v1alpha1"
	"github.com/plantonhq/planton/pkg/cloudflare/r2"
	barmancloudv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/cloudnativepg/kubernetes/barmancloud/v1"
	postgresqlv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/cloudnativepg/kubernetes/postgresql/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createObjectStores renders the Barman Cloud ObjectStore resources:
//
//   - the BACKUP store (named after the cluster) when spec.backup is set —
//     the Cluster's plugins entry points the WAL archiver at it;
//   - the RECOVERY-SOURCE store (`<name>-recovery-source`) when the
//     bootstrap restores from a backup — the synthetic externalClusters
//     entry points at it.
//
// Each store's declared credentials are first materialized as a Secret
// (`<name>-backup-creds` / `<name>-recovery-creds`, plus `-endpoint-ca`
// for self-signed S3-compatible endpoints); keyless arms render the
// backend's ambient-identity flag instead and need no Secret at all.
//
// The backup store is also returned on its own: its UID names the backup
// series of this install (see backupServerName).
func createObjectStores(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependencies []pulumi.ResourceOption,
) ([]pulumi.Resource, *barmancloudv1.ObjectStore, error) {
	var created []pulumi.Resource
	var backupStore *barmancloudv1.ObjectStore

	if backup := locals.Spec.GetBackup(); backup != nil {
		store, err := createObjectStore(ctx, locals, kubernetesProvider, dependencies,
			locals.BackupObjectStoreName, backup.GetObjectStore(), backup.GetRetentionPolicy(),
			locals.BackupCredsSecretName, locals.BackupEndpointCaName)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create backup object store")
		}
		created = append(created, store)
		backupStore = store
	}

	if recovery := locals.Spec.GetBootstrap().GetRecovery(); recovery != nil {
		// Recovery reads an EXISTING archive: retention never applies to it
		// (the plugin must not prune the source cluster's backups).
		store, err := createObjectStore(ctx, locals, kubernetesProvider, dependencies,
			locals.RecoveryObjectStoreName, recovery.GetObjectStore(), "",
			locals.RecoveryCredsSecretName, locals.RecoveryEndpointCaName)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to create recovery-source object store")
		}
		created = append(created, store)
	}

	return created, backupStore, nil
}

// backupServerName resolves the backup SERIES: the folder beneath the
// destination path that this install's base backups and WAL are filed
// under (Barman's server name). A declared backup.server_name is used as
// is. Otherwise the series is `<cluster>-<first 8 characters of the backup
// ObjectStore's UID>`: the ObjectStore is created with the install, before
// the Cluster, so a destroy and recreate from the same declaration
// archives into a fresh series, and an imported cluster keeps its
// ObjectStore and so its series. Barman refuses to archive into a series
// that holds another PostgreSQL system's history, and the refusal is
// quiet (archiving stops, the instances stay healthy, WAL fills the
// volume), which is why the series is never simply the cluster's name.
// The Terraform twin is local.backup_server_name.
func backupServerName(locals *Locals, backupStore *barmancloudv1.ObjectStore) pulumi.StringOutput {
	if declared := locals.Spec.GetBackup().GetServerName(); declared != "" {
		return pulumi.String(declared).ToStringOutput()
	}
	return backupStore.Metadata.Uid().ApplyT(func(uid *string) string {
		short := ""
		if uid != nil {
			short = *uid
		}
		if len(short) > 8 {
			short = short[:8]
		}
		return locals.ClusterName + "-" + short
	}).(pulumi.StringOutput)
}

// createObjectStore renders one ObjectStore resource plus its credential
// satellites. The ObjectStore CRD forbids configuration.serverName (the
// plugin takes it as a per-Cluster parameter instead), so this function
// never sets it.
func createObjectStore(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependencies []pulumi.ResourceOption,
	storeName string,
	objectStore *kubernetespostgresv1alpha1.KubernetesPostgresObjectStore,
	retentionPolicy string,
	credsSecretName string,
	endpointCaSecretName string,
) (*barmancloudv1.ObjectStore, error) {
	configuration := barmancloudv1.ObjectStoreSpecConfigurationArgs{
		DestinationPath: pulumi.String(objectStore.GetDestinationPath()),
	}

	var storeDeps []pulumi.Resource
	// Set by an arm that needs the plugin's sidecar tuned for its store
	// (today: the r2 arm's checksum posture); nil renders no block.
	var instanceSidecarConfiguration *barmancloudv1.ObjectStoreSpecInstanceSidecarConfigurationArgs

	switch {
	case objectStore.GetS3() != nil:
		s3 := objectStore.GetS3()
		s3Credentials := barmancloudv1.ObjectStoreSpecConfigurationS3CredentialsArgs{}
		if s3.GetKeyless() {
			s3Credentials.InheritFromIAMRole = pulumi.Bool(true)
		} else if accessKeys := s3.GetAccessKeys(); accessKeys != nil {
			// Both keys live in one deterministic Secret; the barman config
			// addresses each by key name.
			credsSecret, err := createOpaqueSecret(ctx, locals, kubernetesProvider, dependencies,
				credsSecretName, map[string]string{
					"ACCESS_KEY_ID":     accessKeys.GetAccessKeyId(),
					"SECRET_ACCESS_KEY": accessKeys.GetSecretAccessKey(),
				})
			if err != nil {
				return nil, errors.Wrap(err, "failed to create s3 credentials secret")
			}
			storeDeps = append(storeDeps, credsSecret)
			s3Credentials.AccessKeyId = barmancloudv1.ObjectStoreSpecConfigurationS3CredentialsAccessKeyIdArgs{
				Name: pulumi.String(credsSecretName),
				Key:  pulumi.String("ACCESS_KEY_ID"),
			}
			s3Credentials.SecretAccessKey = barmancloudv1.ObjectStoreSpecConfigurationS3CredentialsSecretAccessKeyArgs{
				Name: pulumi.String(credsSecretName),
				Key:  pulumi.String("SECRET_ACCESS_KEY"),
			}
		}
		configuration.S3Credentials = s3Credentials
		if s3.GetRegion() != "" {
			// The CRD models the region as a SecretKeySelector (not a plain
			// string), so the literal region rides its own deterministic
			// single-key Secret. barman-cloud reads it ONLY on the declared-key
			// path (its credential resolver returns before the region on
			// inheritFromIAMRole, and the AWS SDK derives the region from the
			// pod's IRSA environment there), so on a keyless store this Secret
			// is rendered but inert -- kept for a uniform shape, not for effect.
			regionSecretName := storeName + "-region"
			regionSecret, err := createOpaqueSecret(ctx, locals, kubernetesProvider, dependencies,
				regionSecretName, map[string]string{"AWS_REGION": s3.GetRegion()})
			if err != nil {
				return nil, errors.Wrap(err, "failed to create s3 region secret")
			}
			storeDeps = append(storeDeps, regionSecret)
			s3Credentials.Region = barmancloudv1.ObjectStoreSpecConfigurationS3CredentialsRegionArgs{
				Name: pulumi.String(regionSecretName),
				Key:  pulumi.String("AWS_REGION"),
			}
			configuration.S3Credentials = s3Credentials
		}
		if s3.GetEndpointUrl() != "" {
			configuration.EndpointURL = pulumi.String(s3.GetEndpointUrl())
		}
		if s3.GetEndpointCaPem() != "" {
			caSecret, err := createOpaqueSecret(ctx, locals, kubernetesProvider, dependencies,
				endpointCaSecretName, map[string]string{"ca.crt": s3.GetEndpointCaPem()})
			if err != nil {
				return nil, errors.Wrap(err, "failed to create endpoint CA secret")
			}
			storeDeps = append(storeDeps, caSecret)
			configuration.EndpointCA = barmancloudv1.ObjectStoreSpecConfigurationEndpointCAArgs{
				Name: pulumi.String(endpointCaSecretName),
				Key:  pulumi.String("ca.crt"),
			}
		}

	case objectStore.GetR2() != nil:
		// Cloudflare R2 rides Barman Cloud's S3 code path, but the spec is in
		// R2's own vocabulary; this branch performs the translation:
		//   - endpoint: the jurisdiction's host, composed by the shared
		//     pkg/cloudflare/r2 helper (an eu/fedramp/us bucket is served ONLY
		//     through <account>.<jurisdiction>.r2.cloudflarestorage.com);
		//   - region: "auto", the only region R2 accepts (Cloudflare aliases
		//     "" and us-east-1 to it, but the plugin needs an explicit value
		//     to set AWS_DEFAULT_REGION, so it is always rendered);
		//   - credentials: the S3 key pair as declared -- by default the
		//     CloudflareAccountApiToken's r2_access_key_id/r2_secret_access_key
		//     outputs, which that kind derives (id, SHA-256 of the value); no
		//     hashing happens here. There is no keyless posture for R2.
		// The two AWS_*_CHECKSUM_* variables ride the sidecar for this arm:
		// barman-cloud's boto3 attaches data-integrity checksums by default,
		// which S3-compatible stores may reject; "when_required" is the
		// plugin's documented posture for them and costs nothing on a store
		// that accepts the checksums.
		r2Store := objectStore.GetR2()
		r2Creds := r2Store.GetCredentials()
		credsSecret, err := createOpaqueSecret(ctx, locals, kubernetesProvider, dependencies,
			credsSecretName, map[string]string{
				"ACCESS_KEY_ID":     r2Creds.GetAccessKeyId().GetValue(),
				"SECRET_ACCESS_KEY": r2Creds.GetSecretAccessKey().GetValue(),
			})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create r2 credentials secret")
		}
		storeDeps = append(storeDeps, credsSecret)
		regionSecretName := storeName + "-region"
		regionSecret, err := createOpaqueSecret(ctx, locals, kubernetesProvider, dependencies,
			regionSecretName, map[string]string{"AWS_REGION": r2.Region})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create r2 region secret")
		}
		storeDeps = append(storeDeps, regionSecret)
		configuration.S3Credentials = barmancloudv1.ObjectStoreSpecConfigurationS3CredentialsArgs{
			AccessKeyId: barmancloudv1.ObjectStoreSpecConfigurationS3CredentialsAccessKeyIdArgs{
				Name: pulumi.String(credsSecretName),
				Key:  pulumi.String("ACCESS_KEY_ID"),
			},
			SecretAccessKey: barmancloudv1.ObjectStoreSpecConfigurationS3CredentialsSecretAccessKeyArgs{
				Name: pulumi.String(credsSecretName),
				Key:  pulumi.String("SECRET_ACCESS_KEY"),
			},
			Region: barmancloudv1.ObjectStoreSpecConfigurationS3CredentialsRegionArgs{
				Name: pulumi.String(regionSecretName),
				Key:  pulumi.String("AWS_REGION"),
			},
		}
		configuration.EndpointURL = pulumi.String(r2.S3Endpoint(
			r2Store.GetAccountId().GetValue(), r2Store.GetJurisdiction().GetValue()))
		instanceSidecarConfiguration = &barmancloudv1.ObjectStoreSpecInstanceSidecarConfigurationArgs{
			Env: barmancloudv1.ObjectStoreSpecInstanceSidecarConfigurationEnvArray{
				barmancloudv1.ObjectStoreSpecInstanceSidecarConfigurationEnvArgs{
					Name:  pulumi.String("AWS_REQUEST_CHECKSUM_CALCULATION"),
					Value: pulumi.String("when_required"),
				},
				barmancloudv1.ObjectStoreSpecInstanceSidecarConfigurationEnvArgs{
					Name:  pulumi.String("AWS_RESPONSE_CHECKSUM_VALIDATION"),
					Value: pulumi.String("when_required"),
				},
			},
		}

	case objectStore.GetGcs() != nil:
		gcs := objectStore.GetGcs()
		googleCredentials := barmancloudv1.ObjectStoreSpecConfigurationGoogleCredentialsArgs{}
		if gcs.GetKeyless() {
			googleCredentials.GkeEnvironment = pulumi.Bool(true)
		} else {
			credsSecret, err := createOpaqueSecret(ctx, locals, kubernetesProvider, dependencies,
				credsSecretName, map[string]string{
					"APPLICATION_CREDENTIALS": gcs.GetServiceAccountKeyJson(),
				})
			if err != nil {
				return nil, errors.Wrap(err, "failed to create gcs credentials secret")
			}
			storeDeps = append(storeDeps, credsSecret)
			googleCredentials.ApplicationCredentials = barmancloudv1.ObjectStoreSpecConfigurationGoogleCredentialsApplicationCredentialsArgs{
				Name: pulumi.String(credsSecretName),
				Key:  pulumi.String("APPLICATION_CREDENTIALS"),
			}
		}
		configuration.GoogleCredentials = googleCredentials

	case objectStore.GetAzureBlob() != nil:
		azure := objectStore.GetAzureBlob()
		azureCredentials := barmancloudv1.ObjectStoreSpecConfigurationAzureCredentialsArgs{}
		secretData := map[string]string{}
		switch {
		case azure.GetKeyless():
			azureCredentials.InheritFromAzureAD = pulumi.Bool(true)
			// The storage account still identifies the endpoint; barman
			// reads it from a secret reference even in the keyless posture.
			secretData["AZURE_STORAGE_ACCOUNT"] = azure.GetStorageAccount()
		case azure.GetConnectionString() != "":
			secretData["AZURE_STORAGE_CONNECTION_STRING"] = azure.GetConnectionString()
		default:
			secretData["AZURE_STORAGE_ACCOUNT"] = azure.GetStorageAccount()
			secretData["AZURE_STORAGE_KEY"] = azure.GetStorageKey()
		}
		credsSecret, err := createOpaqueSecret(ctx, locals, kubernetesProvider, dependencies,
			credsSecretName, secretData)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create azure credentials secret")
		}
		storeDeps = append(storeDeps, credsSecret)
		if _, ok := secretData["AZURE_STORAGE_CONNECTION_STRING"]; ok {
			azureCredentials.ConnectionString = barmancloudv1.ObjectStoreSpecConfigurationAzureCredentialsConnectionStringArgs{
				Name: pulumi.String(credsSecretName),
				Key:  pulumi.String("AZURE_STORAGE_CONNECTION_STRING"),
			}
		}
		if _, ok := secretData["AZURE_STORAGE_ACCOUNT"]; ok {
			azureCredentials.StorageAccount = barmancloudv1.ObjectStoreSpecConfigurationAzureCredentialsStorageAccountArgs{
				Name: pulumi.String(credsSecretName),
				Key:  pulumi.String("AZURE_STORAGE_ACCOUNT"),
			}
		}
		if _, ok := secretData["AZURE_STORAGE_KEY"]; ok {
			azureCredentials.StorageKey = barmancloudv1.ObjectStoreSpecConfigurationAzureCredentialsStorageKeyArgs{
				Name: pulumi.String(credsSecretName),
				Key:  pulumi.String("AZURE_STORAGE_KEY"),
			}
		}
		configuration.AzureCredentials = azureCredentials
	}

	if wal := objectStore.GetWal(); wal != nil {
		walArgs := barmancloudv1.ObjectStoreSpecConfigurationWalArgs{}
		if wal.GetCompression() != "" {
			walArgs.Compression = pulumi.String(wal.GetCompression())
		}
		if wal.MaxParallel != nil {
			walArgs.MaxParallel = pulumi.Int(int(wal.GetMaxParallel()))
		}
		configuration.Wal = walArgs
	}

	if data := objectStore.GetData(); data != nil {
		dataArgs := barmancloudv1.ObjectStoreSpecConfigurationDataArgs{}
		if data.GetCompression() != "" {
			dataArgs.Compression = pulumi.String(data.GetCompression())
		}
		if data.Jobs != nil {
			dataArgs.Jobs = pulumi.Int(int(data.GetJobs()))
		}
		if data.GetImmediateCheckpoint() {
			dataArgs.ImmediateCheckpoint = pulumi.Bool(true)
		}
		configuration.Data = dataArgs
	}

	storeSpec := barmancloudv1.ObjectStoreSpecArgs{
		Configuration: configuration,
	}
	if instanceSidecarConfiguration != nil {
		storeSpec.InstanceSidecarConfiguration = instanceSidecarConfiguration
	}
	if retentionPolicy != "" {
		storeSpec.RetentionPolicy = pulumi.String(retentionPolicy)
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependencies...)
	if len(storeDeps) > 0 {
		opts = append(opts, pulumi.DependsOn(storeDeps))
	}

	return barmancloudv1.NewObjectStore(ctx, storeName,
		&barmancloudv1.ObjectStoreArgs{
			Metadata: kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(storeName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: storeSpec,
		}, opts...)
}

// createSeriesStartBackup renders the on-demand Backup every backup series
// starts from (`<cluster>-series-start`). WAL without a base backup cannot
// be replayed, so a series is restorable only from its first base backup
// on; taking one when the series is born makes it restorable from its
// first minute, whatever the schedules say. The series rides the Backup as
// an annotation, and any change replaces the Backup, so a new series (a
// recreate, an upgrade onto per-install series, a changed server_name)
// takes a new base backup, while an unchanged one never re-runs. Deleting
// a Backup resource leaves its stored objects in the bucket; the
// retention policy prunes them.
func createSeriesStartBackup(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependencies []pulumi.ResourceOption,
	backupSeries pulumi.StringOutput,
) error {
	if locals.Spec.GetBackup() == nil {
		return nil
	}

	opts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.ReplaceOnChanges([]string{"*"}),
		pulumi.DeleteBeforeReplace(true),
	}, dependencies...)

	if _, err := postgresqlv1.NewBackup(ctx, locals.SeriesStartBackupName,
		&postgresqlv1.BackupArgs{
			Metadata: kubernetesmeta.ObjectMetaArgs{
				Name:        pulumi.String(locals.SeriesStartBackupName),
				Namespace:   pulumi.String(locals.Namespace),
				Labels:      pulumi.ToStringMap(locals.Labels),
				Annotations: pulumi.StringMap{vars.BackupSeriesAnnotationKey: backupSeries},
			},
			Spec: postgresqlv1.BackupSpecArgs{
				Cluster: postgresqlv1.BackupSpecClusterArgs{
					Name: pulumi.String(locals.ClusterName),
				},
				Method: pulumi.String("plugin"),
				PluginConfiguration: postgresqlv1.BackupSpecPluginConfigurationArgs{
					Name: pulumi.String(vars.BarmanCloudPluginName),
					Parameters: pulumi.StringMap{
						"barmanObjectName": pulumi.String(locals.BackupObjectStoreName),
					},
				},
			},
		}, opts...); err != nil {
		return errors.Wrap(err, "failed to create the series-start backup")
	}
	return nil
}

// createScheduledBackups renders one ScheduledBackup per declared schedule,
// each explicitly method=plugin against the cluster's backup ObjectStore —
// never the deprecated in-tree barmanObjectStore method.
func createScheduledBackups(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependencies []pulumi.ResourceOption,
) error {
	backup := locals.Spec.GetBackup()
	if backup == nil {
		return nil
	}

	for _, schedule := range backup.GetSchedules() {
		scheduledBackupName := locals.ClusterName + "-" + schedule.GetName()

		spec := postgresqlv1.ScheduledBackupSpecArgs{
			Schedule: pulumi.String(schedule.GetSchedule()),
			Cluster: postgresqlv1.ScheduledBackupSpecClusterArgs{
				Name: pulumi.String(locals.ClusterName),
			},
			Method: pulumi.String("plugin"),
			PluginConfiguration: postgresqlv1.ScheduledBackupSpecPluginConfigurationArgs{
				Name: pulumi.String(vars.BarmanCloudPluginName),
				Parameters: pulumi.StringMap{
					"barmanObjectName": pulumi.String(locals.BackupObjectStoreName),
				},
			},
			// Scheduled backups belong to their schedule: deleting the
			// schedule garbage-collects its Backup records while the stored
			// objects in the bucket survive either way.
			BackupOwnerReference: pulumi.String("self"),
		}
		if schedule.GetImmediate() {
			spec.Immediate = pulumi.Bool(true)
		}
		if schedule.GetSuspend() {
			spec.Suspend = pulumi.Bool(true)
		}
		if schedule.GetTarget() != "" {
			spec.Target = pulumi.String(schedule.GetTarget())
		}

		if _, err := postgresqlv1.NewScheduledBackup(ctx, scheduledBackupName,
			&postgresqlv1.ScheduledBackupArgs{
				Metadata: kubernetesmeta.ObjectMetaArgs{
					Name:      pulumi.String(scheduledBackupName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				Spec: spec,
			}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependencies...)...); err != nil {
			return errors.Wrapf(err, "failed to create scheduled backup %s", scheduledBackupName)
		}
	}

	return nil
}

// createOpaqueSecret creates a plain Opaque Secret with the given string
// data — the shape barman's SecretKeySelector references expect.
func createOpaqueSecret(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependencies []pulumi.ResourceOption,
	secretName string, data map[string]string,
) (pulumi.Resource, error) {
	return kubernetescorev1.NewSecret(ctx, secretName,
		&kubernetescorev1.SecretArgs{
			Metadata: kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(secretName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			StringData: pulumi.ToStringMap(data),
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependencies...)...)
}
