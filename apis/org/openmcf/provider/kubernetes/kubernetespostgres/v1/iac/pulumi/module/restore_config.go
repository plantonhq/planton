package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetespostgresv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetespostgres/v1"
	zalandov1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/zalandooperator/kubernetes/acid/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildRestoreStandbyBlock generates the Zalando operator's spec.standby block for
// cross-cluster disaster recovery using the Standby-then-Promote pattern.
//
// When restore is enabled it returns a populated standby block with s3_wal_path so the
// database bootstraps read-only from the backup WAL path. When restore is disabled or
// nil it returns nil (normal read-write primary; removing a previously-set standby
// triggers promotion).
func buildRestoreStandbyBlock(
	restoreConfig *kubernetespostgresv1.KubernetesPostgresRestoreConfig,
) (*zalandov1.PostgresqlSpecStandbyArgs, error) {
	if restoreConfig == nil || !restoreConfig.Enabled {
		return nil, nil
	}

	if restoreConfig.S3Path == "" {
		return nil, errors.New("restore_config.s3_path is required when restore_config.enabled=true")
	}
	if restoreConfig.BucketName == nil || *restoreConfig.BucketName == "" {
		return nil, errors.New("restore_config.bucket_name is required when restore_config.enabled=true")
	}

	// Format: s3://bucket-name/path/to/backups
	fullS3Path := fmt.Sprintf("s3://%s/%s", *restoreConfig.BucketName, restoreConfig.S3Path)
	return &zalandov1.PostgresqlSpecStandbyArgs{
		S3_wal_path: pulumi.String(fullS3Path),
	}, nil
}

// buildRestoreEnvVars returns the STANDBY_* spec.env entries used by Spilo/Patroni
// during standby bootstrap. Endpoint and path-style are plain values; the credentials
// are injected via a generated Secret + secretKeyRef (never plaintext). Returns nil
// when restore is disabled or no r2_config is supplied.
func buildRestoreEnvVars(
	ctx *pulumi.Context,
	kubernetesProvider pulumi.ProviderResource,
	locals *Locals,
	namespaceDeps []pulumi.ResourceOption,
	restoreConfig *kubernetespostgresv1.KubernetesPostgresRestoreConfig,
) ([]pulumi.MapInput, error) {
	if restoreConfig == nil || !restoreConfig.Enabled || restoreConfig.R2Config == nil {
		return nil, nil
	}

	r2 := restoreConfig.R2Config
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r2.CloudflareAccountId)

	// STANDBY_* env vars are used by Spilo during standby bootstrap; these are
	// distinct from the WALG_*/AWS_* (ongoing backup) set.
	envVars := []pulumi.MapInput{
		envVar("STANDBY_AWS_ENDPOINT", endpoint),
		envVar("STANDBY_AWS_FORCE_PATH_STYLE", "true"),
	}

	credEnvVars, err := r2CredentialEnvVars(ctx, kubernetesProvider, locals.Namespace, namespaceDeps,
		fmt.Sprintf("%s-restore-r2-credentials", locals.KubernetesPostgres.Metadata.Name),
		"STANDBY_AWS_ACCESS_KEY_ID", "STANDBY_AWS_SECRET_ACCESS_KEY",
		locals.Labels, r2.AccessKeyId, r2.SecretAccessKey)
	if err != nil {
		return nil, err
	}
	envVars = append(envVars, credEnvVars...)

	return envVars, nil
}
