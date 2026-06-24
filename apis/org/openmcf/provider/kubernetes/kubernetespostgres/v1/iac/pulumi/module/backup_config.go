package module

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	kubernetespostgresv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetespostgres/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// envVar builds a single name/value entry for the Zalando postgresql spec.env list.
func envVar(name, value string) pulumi.MapInput {
	return pulumi.Map{
		"name":  pulumi.String(name),
		"value": pulumi.String(value),
	}
}

// r2CredentialEnvVars creates a Kubernetes Secret holding the R2 access-key id and
// secret access key, then returns two spec.env entries that reference those keys via
// secretKeyRef. The credentials therefore never appear in plaintext in the postgresql
// custom resource or the pod spec. The env var NAMES are parameterized so the same
// helper serves both the backup (AWS_*) and restore (STANDBY_AWS_*) directions.
func r2CredentialEnvVars(
	ctx *pulumi.Context,
	kubernetesProvider pulumi.ProviderResource,
	namespace string,
	namespaceDeps []pulumi.ResourceOption,
	secretName string,
	accessKeyEnvName string,
	secretKeyEnvName string,
	labels map[string]string,
	accessKeyId string,
	secretAccessKey string,
) ([]pulumi.MapInput, error) {
	const (
		accessKeyDataKey = "access_key_id"
		secretKeyDataKey = "secret_access_key"
	)

	secretOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := corev1.NewSecret(ctx,
		secretName,
		&corev1.SecretArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(secretName),
				Namespace: pulumi.String(namespace),
				Labels:    pulumi.ToStringMap(labels),
			},
			Type: pulumi.String("Opaque"),
			StringData: pulumi.StringMap{
				accessKeyDataKey: pulumi.String(accessKeyId),
				secretKeyDataKey: pulumi.String(secretAccessKey),
			},
		},
		secretOpts...,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create r2 credentials secret %s", secretName)
	}

	return []pulumi.MapInput{
		pulumi.Map{
			"name": pulumi.String(accessKeyEnvName),
			"valueFrom": pulumi.Map{
				"secretKeyRef": pulumi.Map{
					"name": pulumi.String(secretName),
					"key":  pulumi.String(accessKeyDataKey),
				},
			},
		},
		pulumi.Map{
			"name": pulumi.String(secretKeyEnvName),
			"valueFrom": pulumi.Map{
				"secretKeyRef": pulumi.Map{
					"name": pulumi.String(secretName),
					"key":  pulumi.String(secretKeyDataKey),
				},
			},
		},
	}, nil
}

// buildBackupEnvVars returns the spec.env entries that configure per-database WAL-G
// backups. The WAL-G push target is composed from the bucket and object_prefix; when
// credentials are supplied the dedicated R2 endpoint and the credentials (via a
// generated Secret + secretKeyRef) are wired so the database backs up to its own
// bucket. Returns nil when backups are disabled.
func buildBackupEnvVars(
	ctx *pulumi.Context,
	kubernetesProvider pulumi.ProviderResource,
	locals *Locals,
	namespaceDeps []pulumi.ResourceOption,
	backupConfig *kubernetespostgresv1.KubernetesPostgresBackupConfig,
) ([]pulumi.MapInput, error) {
	if backupConfig == nil || !backupConfig.Enabled {
		return nil, nil
	}

	envVars := []pulumi.MapInput{envVar("USE_WALG_BACKUP", "true")}

	if walgS3Prefix := backupWalgS3Prefix(backupConfig); walgS3Prefix != "" {
		envVars = append(envVars, envVar("WALG_S3_PREFIX", walgS3Prefix))
	}

	if backupConfig.Schedule != "" {
		envVars = append(envVars, envVar("BACKUP_SCHEDULE", backupConfig.Schedule))
	}

	if backupConfig.RetainCount > 0 {
		envVars = append(envVars, envVar("BACKUP_NUM_TO_RETAIN", strconv.Itoa(int(backupConfig.RetainCount))))
	}

	// Dedicated R2 backup target: endpoint/region/path-style as plain values and the
	// credentials via a generated Secret + secretKeyRef. USE_WALG_RESTORE is enabled
	// so WAL-G can also fetch from the same bucket when needed.
	if creds := backupConfig.Credentials; creds != nil {
		endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", creds.CloudflareAccountId)
		envVars = append(envVars,
			envVar("AWS_ENDPOINT", endpoint),
			envVar("AWS_REGION", "auto"),
			envVar("AWS_FORCE_PATH_STYLE", "true"),
			envVar("USE_WALG_RESTORE", "true"),
		)

		credEnvVars, err := r2CredentialEnvVars(ctx, kubernetesProvider, locals.Namespace, namespaceDeps,
			fmt.Sprintf("%s-backup-r2-credentials", locals.KubernetesPostgres.Metadata.Name),
			"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY",
			locals.Labels, creds.AccessKeyId, creds.SecretAccessKey)
		if err != nil {
			return nil, err
		}
		envVars = append(envVars, credEnvVars...)
	}

	return envVars, nil
}

// backupWalgS3Prefix composes the WAL-G push target as
// s3://<bucket>[/<object_prefix>]/$(SCOPE)/$(PGVERSION). Spilo/Patroni substitutes
// the $(SCOPE)/$(PGVERSION) suffix at runtime, so backups from many clusters and
// versions share one bucket without collision. Returns "" when no bucket is set.
func backupWalgS3Prefix(backupConfig *kubernetespostgresv1.KubernetesPostgresBackupConfig) string {
	bucket := backupConfig.GetBucket().GetValue()
	if bucket == "" {
		return ""
	}
	prefix := fmt.Sprintf("s3://%s", bucket)
	if objectPrefix := backupConfig.ObjectPrefix; objectPrefix != "" {
		prefix = fmt.Sprintf("%s/%s", prefix, objectPrefix)
	}
	return fmt.Sprintf("%s/$(SCOPE)/$(PGVERSION)", prefix)
}
