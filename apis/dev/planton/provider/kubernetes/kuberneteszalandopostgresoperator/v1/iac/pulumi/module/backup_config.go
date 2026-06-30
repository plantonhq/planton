package module

import (
	"fmt"

	"github.com/pkg/errors"
	kuberneteszalandopostgresoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteszalandopostgresoperator/v1"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createBackupResources creates the Secret and ConfigMap for PostgreSQL backups using Cloudflare R2.
// Returns the ConfigMap name if backup is configured, empty string otherwise.
func createBackupResources(
	ctx *pulumi.Context,
	locals *Locals,
	backupConfig *kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperatorBackupConfig,
	namespace string,
	kubernetesProvider *pulumikubernetes.Provider,
	labels map[string]string,
	namespaceDeps []pulumi.ResourceOption,
) (pulumi.StringOutput, error) {
	// If no backup config provided, return empty
	if backupConfig == nil {
		return pulumi.String("").ToStringOutput(), nil
	}

	creds := backupConfig.Credentials
	if creds == nil {
		return pulumi.String("").ToStringOutput(), errors.New("backup_config.credentials is required when backup_config is specified")
	}

	// 1. Create Secret for R2 credentials
	secretOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	createdSecret, err := corev1.NewSecret(ctx,
		locals.BackupSecretName,
		&corev1.SecretArgs{
			Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.BackupSecretName),
				Namespace: pulumi.String(namespace),
				Labels:    pulumi.ToStringMap(labels),
			}),
			Type: pulumi.String("Opaque"),
			StringData: pulumi.StringMap{
				"AWS_ACCESS_KEY_ID":     pulumi.String(creds.AccessKeyId),
				"AWS_SECRET_ACCESS_KEY": pulumi.String(creds.SecretAccessKey),
			},
		},
		secretOpts...,
	)
	if err != nil {
		return pulumi.String("").ToStringOutput(), errors.Wrap(err, "failed to create backup credentials secret")
	}

	// 2. Build R2 endpoint URL
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", creds.CloudflareAccountId)

	// 3. Build ConfigMap data
	configMapData := pulumi.StringMap{
		// WAL-G flags (default to true if not explicitly disabled)
		"USE_WALG_BACKUP":        pulumi.String(boolToString(backupConfig.EnableWalGBackup, true)),
		"USE_WALG_RESTORE":       pulumi.String(boolToString(backupConfig.EnableWalGRestore, true)),
		"CLONE_USE_WALG_RESTORE": pulumi.String(boolToString(backupConfig.EnableCloneWalGRestore, true)),

		// S3/R2 configuration
		"WALG_S3_PREFIX":       pulumi.String(backupWalgS3Prefix(backupConfig)),
		"AWS_ENDPOINT":         pulumi.String(r2Endpoint),
		"AWS_REGION":           pulumi.String("auto"), // R2 uses "auto" region
		"AWS_FORCE_PATH_STYLE": pulumi.String("true"), // Required for R2

		// Backup schedule
		"BACKUP_SCHEDULE": pulumi.String(backupConfig.Schedule),

		// Credentials (reference the Secret - Zalando operator will mount them)
		// We don't include credentials in ConfigMap; they come from the Secret
		// The Secret will be mounted at /run/etc/wal-e.d/env by Zalando operator
		"AWS_ACCESS_KEY_ID":     pulumi.String(creds.AccessKeyId),
		"AWS_SECRET_ACCESS_KEY": pulumi.String(creds.SecretAccessKey),
	}

	// 5. Create ConfigMap
	cmOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn([]pulumi.Resource{createdSecret}),
	}, namespaceDeps...)
	createdConfigMap, err := corev1.NewConfigMap(ctx,
		locals.BackupConfigMapName,
		&corev1.ConfigMapArgs{
			Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.BackupConfigMapName),
				Namespace: pulumi.String(namespace),
				Labels:    pulumi.ToStringMap(labels),
			}),
			Data: configMapData,
		},
		cmOpts...,
	)
	if err != nil {
		return pulumi.String("").ToStringOutput(), errors.Wrap(err, "failed to create backup config map")
	}

	// Return the ConfigMap name (namespace/name format for Zalando operator)
	return pulumi.Sprintf("%s/%s", namespace, createdConfigMap.Metadata.Name().Elem()), nil
}

// boolToString converts a bool to "true"/"false" string, with a default value when the bool is false.
func boolToString(value bool, defaultWhenFalse bool) string {
	if value {
		return "true"
	}
	if defaultWhenFalse {
		return "true"
	}
	return "false"
}

// backupWalgS3Prefix composes the WAL-G push target as
// s3://<bucket>[/<object_prefix>]/$(SCOPE)/$(PGVERSION). One operator configmap serves
// every database on the cluster, so Spilo/Patroni substitutes the $(SCOPE)/$(PGVERSION)
// suffix per database at runtime.
func backupWalgS3Prefix(backupConfig *kuberneteszalandopostgresoperatorv1.KubernetesZalandoPostgresOperatorBackupConfig) string {
	prefix := fmt.Sprintf("s3://%s", backupConfig.GetBucket().GetValue())
	if objectPrefix := backupConfig.ObjectPrefix; objectPrefix != "" {
		prefix = fmt.Sprintf("%s/%s", prefix, objectPrefix)
	}
	return fmt.Sprintf("%s/$(SCOPE)/$(PGVERSION)", prefix)
}
