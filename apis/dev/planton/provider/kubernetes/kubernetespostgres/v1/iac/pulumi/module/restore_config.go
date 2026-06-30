package module

import (
	"fmt"

	kubernetespostgresv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetespostgres/v1"
	zalandov1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/zalandooperator/kubernetes/acid/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildRestoreStandbyBlock generates the Zalando operator's spec.standby block for
// cross-cluster disaster recovery using the Standby-then-Promote pattern.
//
// When restore is enabled it returns a populated standby block pointing at
// s3://<bucket>/<object_prefix> so the database bootstraps read-only from the source
// backup. When restore is disabled or nil it returns nil (normal read-write primary;
// removing a previously-set standby triggers promotion).
func buildRestoreStandbyBlock(
	restoreConfig *kubernetespostgresv1.KubernetesPostgresRestoreConfig,
) *zalandov1.PostgresqlSpecStandbyArgs {
	if restoreConfig == nil || !restoreConfig.Enabled {
		return nil
	}

	s3WalPath := fmt.Sprintf("s3://%s/%s", restoreConfig.GetBucket().GetValue(), restoreConfig.ObjectPrefix)
	return &zalandov1.PostgresqlSpecStandbyArgs{
		S3_wal_path: pulumi.String(s3WalPath),
	}
}

// buildRestoreEnvVars returns the STANDBY_* spec.env entries used by Spilo/Patroni
// during standby bootstrap. Endpoint and path-style are plain values; the credentials
// are injected via a generated Secret + secretKeyRef (never plaintext). Returns nil
// when restore is disabled or no credentials are supplied.
func buildRestoreEnvVars(
	ctx *pulumi.Context,
	kubernetesProvider pulumi.ProviderResource,
	locals *Locals,
	namespaceDeps []pulumi.ResourceOption,
	restoreConfig *kubernetespostgresv1.KubernetesPostgresRestoreConfig,
) ([]pulumi.MapInput, error) {
	if restoreConfig == nil || !restoreConfig.Enabled || restoreConfig.Credentials == nil {
		return nil, nil
	}

	creds := restoreConfig.Credentials
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", creds.CloudflareAccountId)

	// STANDBY_* env vars are used by Spilo during standby bootstrap; these are
	// distinct from the WALG_*/AWS_* (ongoing backup) set.
	envVars := []pulumi.MapInput{
		envVar("STANDBY_AWS_ENDPOINT", endpoint),
		envVar("STANDBY_AWS_FORCE_PATH_STYLE", "true"),
	}

	credEnvVars, err := r2CredentialEnvVars(ctx, kubernetesProvider, locals.Namespace, namespaceDeps,
		fmt.Sprintf("%s-restore-r2-credentials", locals.KubernetesPostgres.Metadata.Name),
		"STANDBY_AWS_ACCESS_KEY_ID", "STANDBY_AWS_SECRET_ACCESS_KEY",
		locals.Labels, creds.AccessKeyId, creds.SecretAccessKey)
	if err != nil {
		return nil, err
	}
	envVars = append(envVars, credEnvVars...)

	return envVars, nil
}
