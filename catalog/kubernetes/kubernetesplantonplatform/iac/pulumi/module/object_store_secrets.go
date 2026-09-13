package module

import (
	"github.com/pkg/errors"
	kubernetesplantonplatformv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesplantonplatform/v1alpha1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The keys the operator reads from an object store's credentials Secret,
// per backend — its contract, mirrored here so the Secret this module
// writes is the Secret the operator's preflight looks for. S3 and R2 share
// the S3 key pair (R2's pair is derived from a Cloudflare API token); GCS
// carries a service-account JSON key; Azure Blob a connection string.
const (
	objectStoreKeyAccessKeyID            = "ACCESS_KEY_ID"
	objectStoreKeySecretAccessKey        = "SECRET_ACCESS_KEY"
	objectStoreKeyApplicationCredentials = "APPLICATION_CREDENTIALS"
	objectStoreKeyAzureConnectionString  = "AZURE_STORAGE_CONNECTION_STRING"
)

// objectStoreSecret is one Secret the declaration needs the module to
// materialize: the credential a store's arm declares, or the CA bundle a
// private S3 endpoint chains to. The spec speaks in declared values (and
// references resolved to values before this module runs); the operator
// speaks in Secret names; these are the translation.
type objectStoreSecret struct {
	Name string
	Data map[string]string
}

// objectStoreSecrets lists every Secret the platform's backup and recovery
// declarations need, in the order they are created. Empty when neither is
// declared, or when every declared store is keyless — a keyless posture
// names no Secret and the CR carries no credentialsSecretName for it.
// Pure so the rendering can be tested without a Pulumi runtime.
func objectStoreSecrets(locals *Locals) []objectStoreSecret {
	pg := locals.Spec.GetDatabase().GetPostgresql()
	if pg == nil {
		return nil
	}
	var out []objectStoreSecret
	if b := pg.GetBackup(); b != nil {
		out = append(out, storeSecrets(b.GetObjectStore(),
			locals.BackupCredentialsSecretName, locals.BackupEndpointCaSecretName)...)
	}
	if r := pg.GetRecoverFrom(); r != nil {
		out = append(out, storeSecrets(r.GetObjectStore(),
			locals.RecoveryCredentialsSecretName, locals.RecoveryEndpointCaSecretName)...)
	}
	return out
}

// storeSecrets is the Secret set for one store: its credential (when the
// arm declares one) and, for an S3-compatible endpoint with a private CA,
// the CA bundle under the one key the CR's endpointCASecretRef names.
func storeSecrets(store *kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore,
	credentialsSecretName, endpointCaSecretName string) []objectStoreSecret {
	var out []objectStoreSecret
	if data := objectStoreCredentialsData(store); data != nil {
		out = append(out, objectStoreSecret{Name: credentialsSecretName, Data: data})
	}
	if s3 := store.GetS3(); s3 != nil && s3.GetEndpointCaPem() != "" {
		out = append(out, objectStoreSecret{
			Name: endpointCaSecretName,
			Data: map[string]string{vars.EndpointCaSecretKey: s3.GetEndpointCaPem()},
		})
	}
	return out
}

// objectStoreCredentialsData is the credentials Secret's data for a store,
// or nil when the store's posture is keyless (no Secret exists, and the CR
// omits credentialsSecretName so the operator reads the pods' own cloud
// identity). References in the R2 arm arrive already resolved to values.
func objectStoreCredentialsData(store *kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore) map[string]string {
	switch {
	case store.GetS3() != nil:
		keys := store.GetS3().GetAccessKeys()
		if keys == nil {
			return nil
		}
		return map[string]string{
			objectStoreKeyAccessKeyID:     keys.GetAccessKeyId(),
			objectStoreKeySecretAccessKey: keys.GetSecretAccessKey(),
		}
	case store.GetGcs() != nil:
		if store.GetGcs().GetServiceAccountKeyJson() == "" {
			return nil
		}
		return map[string]string{
			objectStoreKeyApplicationCredentials: store.GetGcs().GetServiceAccountKeyJson(),
		}
	case store.GetAzureBlob() != nil:
		if store.GetAzureBlob().GetConnectionString() == "" {
			return nil
		}
		return map[string]string{
			objectStoreKeyAzureConnectionString: store.GetAzureBlob().GetConnectionString(),
		}
	case store.GetR2() != nil:
		creds := store.GetR2().GetCredentials()
		return map[string]string{
			objectStoreKeyAccessKeyID:     creds.GetAccessKeyId().GetValue(),
			objectStoreKeySecretAccessKey: creds.GetSecretAccessKey().GetValue(),
		}
	}
	return nil
}

// createObjectStoreSecrets materializes the Secrets objectStoreSecrets
// lists, after the namespace and before the CR (the caller adds the
// returned resources to the CR's DependsOn). Created first so the database
// is born archiving: the operator holds nothing for a credential that is
// already there. Terraform equivalent: kubernetes_secret_v1 with count on
// the same names.
func createObjectStoreSecrets(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependencies []pulumi.ResourceOption,
) ([]pulumi.Resource, error) {
	var created []pulumi.Resource
	for _, s := range objectStoreSecrets(locals) {
		secret, err := kubernetescorev1.NewSecret(ctx, s.Name,
			&kubernetescorev1.SecretArgs{
				Metadata: kubernetesmeta.ObjectMetaArgs{
					Name:      pulumi.String(s.Name),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				StringData: pulumi.ToStringMap(s.Data),
			}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependencies...)...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create the %s Secret", s.Name)
		}
		created = append(created, secret)
	}
	return created, nil
}
